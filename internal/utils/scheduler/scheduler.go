package scheduler

import (
	"context"
	"github.com/google/uuid"
)

// struct with method run
type Task interface {
	Run(ctx context.Context) error
}

func NewScheduler() Scheduler {
	return Scheduler{}
}

type Scheduler struct {
	tasks     map[string]interface{}
	schedule  map[string]Task
	TaskGroup string
	// active
	// holding
	// Мин интервал - секунда
	// Очередь на создание
	// Не проходить все заявки по циклу

	// бинарное дерево до секунд [1, 20, 56, 78]
	// каждые n секунд запустить со счётчиками 7..12..17..22
	// в определённое время один раз 20:40:03
	// ровно каждые n-секунд 5..10..15..20

	// Как исполнять?
	// Гарантированное исполнение
	// Защита от повторного исполнения
	// Параллельное исполнение (быстро)

	// Рассылка в телеграм
	// Отправка заявок на биржу
}

func (s *Scheduler) AddTask(Task) {

}

func (s *Scheduler) RemoveTask(taskID uuid.UUID) {

}

func (s *Scheduler) Start() {
	// go loop()
	// select
	//
	// if TaskGroup.Start < time.Time {
	// s.ThrowGroup()
	// s.NextGroup()

}

func (s *Scheduler) Stop() {
	// ctx.Cancel()
}

//func (s *Scheduler) NextGroup() {
//	s.TaskGroup = s.TaskGroup.Next()
//}

func (s *Scheduler) ThrowGroup() {
	// ctx.Cancel()
}

type Link struct {
	next *Link
}

func (l *Link) Next() {

}
