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

	menu := &tele.ReplyMarkup{}
	btnAdd := menu.Data("➕ Дбавить запись", "add")
	btnDelete := menu.Data("➖ Удалить последнюю запись", "delete")
	btnGetAvatage := menu.Data("🏆 Узнать среднее", "avarage")
	btnGetSalary := menu.Data("🤑 Узнать ЗП", "salary")
	btnGetTotalMonth := menu.Data("💰 Узнать выручку за месяц", "totalMonth")
	btnGetAllRow := menu.Data("👀 Увидеть все записи за месяц", "allRow")
	menu.Inline(
		menu.Row(btnAdd),
		menu.Row(btnGetSalary),
		menu.Row(btnGetAvatage),
		menu.Row(btnGetTotalMonth),
		menu.Row(btnGetAllRow),
		menu.Row(btnDelete),
	)
	var message *tele.Message
	var stateAdd bool
	ctx := context.Background()
	q := db.New(conn)

	b.Handle("/menu", func(c tele.Context) error {
		m, err := b.Send(tele.ChatID(c.Chat().ID), "Привет, я предоставлю тебе все цифры которые тебе нужны!", menu)
		if err != nil {
			return c.Send("Ууупс... что-то пошло не так: %v", err)
		}
		message = m
		return nil
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

		_, err = b.Edit(message,
			fmt.Sprintf("Запись:\n\n%v: <b>%v</b>\n\nбыла успешно удалена 😉", lastVal.Name, lastVal.Val), menu)
		if err != nil {
			return c.Send("Эта информация уже на экране!")
		}
		return nil
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
		_, err = b.Edit(message, fmt.Sprintf("Твое среднее: <b>%.1f</b>\nСреднее какой-то Дашки: <b>%.1f</b>\n\nПо итогу: <b>%.1f</b>", avgAlena.Float64, avgDasha.Float64, avgAlena.Float64-avgDasha.Float64), menu)
		if err != nil {
			return c.Send("Эта информация уже на экране!")
		}
		return nil
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
		_, err = b.Edit(message,
			fmt.Sprintf("Твоя ЗП на текущий момент: <b>%v</b>\nА было бы: <b>%v</b>",
				result.Count*1500+(int64(result.Sum.Float64*0.04)),
				result.Count*3000,
			), menu)
		if err != nil {
			return c.Send("Эта информация уже на экране!")
		}
		return nil
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
		_, err = b.Edit(message, fmt.Sprintf("Всего в этом месяце: <b>%.1f</b>", r.Float64), menu)
		if err != nil {
			return c.Send("Эта информация уже на экране!")
		}
		return nil
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
		_, err = b.Edit(message, msg.String(), menu)
		if err != nil {
			return c.Send("Эта информация уже на экране!")
		}
		return nil
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		if stateAdd {
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
			m, err := b.Send(tele.ChatID(c.Chat().ID), "Запись была успешно добавлена 😉", menu)
			if err != nil {
				return c.Send(
					fmt.Sprintf("Ууупс... что-то пошло не так: %v", err))

			}
			stateAdd = false
			message = m
			return nil
		}
		return nil
	})

	b.Start()

}
