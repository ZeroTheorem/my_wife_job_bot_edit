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
	// -- Section: load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// -- end section

	// -- Section: create and setup bot object
	b, err := tele.NewBot(tele.Settings{
		Token:     os.Getenv("TOKEN"),
		Poller:    &tele.LongPoller{Timeout: 10 * time.Second},
		ParseMode: tele.ModeHTML,
	})

	if err != nil {
		log.Fatal(err)
	}
	// -- end section

	// -- Section: open db connection and setub query executor
	conn, err := sql.Open("sqlite", "file:mydb.db")

	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	q := db.New(conn)
	// -- end section

	// -- Section: prepare keyboard and buttons
	menu := &tele.ReplyMarkup{}
	btnAdd := menu.Data("➕ Дбавить запись", "add")
	btnDelete := menu.Data("➖ Удалить последнюю запись", "delete")
	btnGetAvatage := menu.Data("🏆 Узнать среднее", "avarage")
	btnGetSalary := menu.Data("🤑 Узнать ЗП", "salary")
	btnGetTotalMonth := menu.Data("💰 Узнать выручку за месяц", "totalMonth")
	btnGetAllRow := menu.Data("👀 Увидеть все записи за месяц", "allRow")
	btnSetTarget := menu.Data("🎯 Установить палан на месяц", "target")
	menu.Inline(
		menu.Row(btnAdd),
		menu.Row(btnGetSalary),
		menu.Row(btnGetAvatage),
		menu.Row(btnGetTotalMonth),
		menu.Row(btnSetTarget),
		menu.Row(btnGetAllRow),
		menu.Row(btnDelete),
	)
	// -- end section

	// -- Section: define states
	var stateAdd bool
	var stateSetTarget bool
	var target int64
	// -- end section

	// -- Section: define hanlers
	b.Handle("/menu", func(c tele.Context) error {
		return c.Send("Привет, я предоставлю тебе все цифры которые тебе нужны!", menu)
	})
	b.Handle(&btnAdd, func(c tele.Context) error {
		stateAdd = true
		return c.Send("Введи сообщение в формате имя|значение")
	})
	b.Handle(&btnDelete, func(c tele.Context) error {
		lastVal, err := q.DeleteLastRow(ctx)
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		return c.Edit(fmt.Sprintf("Запись:\n\n%v: <b>%v</b>\n\nбыла успешно удалена 😉", lastVal.Name, lastVal.Val), menu)

	})

	b.Handle(&btnGetAvatage, func(c tele.Context) error {
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
		return c.Edit(fmt.Sprintf("Твое среднее: <b>%.1f</b>\nСреднее какой-то Дашки: <b>%.1f</b>\n\nПо итогу: <b>%.1f</b>", avgAlena.Float64, avgDasha.Float64, avgAlena.Float64-avgDasha.Float64), menu)
	})

	b.Handle(&btnGetSalary, func(c tele.Context) error {
		result, err := q.GetWifeSalary(ctx, db.GetWifeSalaryParams{
			Name:  "алена",
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		return c.Edit(
			fmt.Sprintf("Твоя ЗП на текущий момент: <b>%v</b>\nА было бы: <b>%v</b>",
				result.Count*1500+(int64(result.Sum.Float64*0.04)),
				result.Count*3000,
			), menu)
	})

	b.Handle(&btnGetTotalMonth, func(c tele.Context) error {
		r, err := q.GetMonthlyTotal(ctx, db.GetMonthlyTotalParams{
			Month: int64(time.Now().Month()),
			Year:  int64(time.Now().Year()),
		})
		if err != nil {
			return c.Send(
				fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))
		}
		return c.Edit(fmt.Sprintf("Всего в этом месяце: <b>%.1f</b>", r.Float64), menu)
	})
	b.Handle(&btnGetAllRow, func(c tele.Context) error {
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
		return c.Edit(msg.String(), menu)
	})

	b.Handle(&btnSetTarget, func(c tele.Context) error {
		stateSetTarget = true
		return c.Send("Введите значение")
	})
	b.Handle(tele.OnText, func(c tele.Context) error {
		switch {
		case stateAdd:
			vals := strings.Split(c.Message().Text, " ")
			if len(vals) != 2 {
				return c.Send("Необходимо ввести все значение в формате имя|начение")
			}
			nameLower := strings.ToLower(vals[0])
			if nameLower != "даша" && nameLower != "алена" {
				return c.Send(
					"Допустимые имена:\n\nДаша\nАлена\n\nможешь писать их с маленькой или большой буквы - это не важно, но другие имена не допустимы!")
			}
			intValue, err := strconv.ParseInt(vals[1], 10, 64)
			if err != nil {
				return c.Send(
					fmt.Sprintf("%v -- второе значение после /add должно быть числом", vals[1]))
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
			stateAdd = false
			return c.Send("Запись была успешно добавлена 😉", menu)
		case stateSetTarget:
			msg := c.Message().Text
			i, err := strconv.ParseInt(msg, 10, 64)
			if err != nil {
				c.Send("Пожалуйста введите число!")
			}
			target = i
			stateSetTarget = false
			return c.Send(fmt.Sprintf("План %v был успешно установлен! 😉", target), menu)
		}
		return nil
	})
	// -- end section

	// -- Section: start app
	b.Start()
	// -- end section

}
