package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ZeroTheorem/my_wife_job_bot/db"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v4"
	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	b, err := tele.NewBot(tele.Settings{
		Token:     os.Getenv("TOKEN"),
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeHTML,
	})

	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open("sqlite", "file:mydb.db")

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	q := db.New(conn)

	b.Handle("/add", func(c tele.Context) error {
		vals := strings.Split(c.Message().Text, " ")
		if len(vals) != 3 {
			return c.Send("Необходимо ввести все значение в формате\n\n/add <Имя> <Значение>")
		}

		nameLower := strings.ToLower(vals[1])
		if nameLower != "даша" && nameLower != "алена" {
			return c.Send(
				"Допустимые имена:\n\nДаша\nАлена\n\nможешь писать их с маленькой или большой буквы - это не важно, но другие имена не допустимы!")
		}

		intValue, err := strconv.ParseInt(vals[2], 10, 64)
		if err != nil {
			return c.Send(
				fmt.Sprintf("%v -- второе значение после /add должно быть числом", vals[2]))
		}

		err = q.CreateRow(ctx, db.CreateRowParams{
			Name:  nameLower,
			Val:   intValue,
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})

		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))

		}

		return c.Send("Запись была успешно добавлена 😉")
	})
	b.Handle("/deletelast", func(c tele.Context) error {
		lastVal, err := q.DeleteLastRow(ctx)

		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}

		return c.Send(
			fmt.Sprintf("Запись:\n\n%v: <b>%v</b>\n\nбыла успешно удалена 😉", lastVal.Name, lastVal.Val))
	})

	b.Handle("/whowins", func(c tele.Context) error {
		avgDasha, err := q.GetAvg(ctx, db.GetAvgParams{
			Name:  "даша",
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		avgAlena, err := q.GetAvg(ctx, db.GetAvgParams{
			Name:  "алена",
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		return c.Send(fmt.Sprintf("Твое среднее: <b>%.1f</b>\nСреднее какой-то Дашки: <b>%.1f</b>\n\nПо итогу: <b>%.1f</b>", avgAlena.Float64, avgDasha.Float64, avgAlena.Float64-avgDasha.Float64))

	})

	b.Handle("/mysalary", func(c tele.Context) error {
		result, err := q.GetWifeSalary(ctx, db.GetWifeSalaryParams{
			Name:  "алена",
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		return c.Send(
			fmt.Sprintf("Твоя ЗП на текущий момент: <b>%v</b>\nА было бы: <b>%v</b>",
				result.Count*1500+(int64(result.Sum.Float64*0.04)),
				result.Count*3000,
			))
	})

	b.Handle("/totalmonth", func(c tele.Context) error {
		r, err := q.GetMonthlyTotal(ctx, db.GetMonthlyTotalParams{
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		return c.Send(fmt.Sprintf("Всего в этом месяце: <b>%.1f</b>", r.Float64))

	})
	b.Handle("/all", func(c tele.Context) error {
		r, err := q.GetAllRowsInMonth(ctx, db.GetAllRowsInMonthParams{
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		var msg strings.Builder
		for _, v := range r {
			fmt.Fprintf(&msg, "%v.%v -- %v: <b>%v</b>\n", v.Month, v.Year, v.Name, v.Val)
		}
		return c.Send(msg.String())
	})

	b.Start()

}
