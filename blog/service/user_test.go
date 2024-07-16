package service

import "testing"

func TestSaveUser(t *testing.T) {
	SaveUser()
}

func TestSaveUserBatch(t *testing.T) {
	SaveUserBatch()
}

func TestUpdateUser(t *testing.T) {
	UpdateUser()
}

func TestUpdateUserByUser(t *testing.T) {
	UpdateUserByUser()
}

func TestUpdateUserByMethod(t *testing.T) {
	UpdateUserByMethod()
}

func TestSelectOne(t *testing.T) {
	SelectOne()
}

func TestWhere(t *testing.T) {
	Where()
}

func TestDelete(t *testing.T) {
	Delete()
}

func TestSelect(t *testing.T) {
	Select()
}

func TestAvg(t *testing.T) {
	Avg()
}

func TestAggregate(t *testing.T) {
	Aggregate()
}

func TestQueryRow(t *testing.T) {
	QueryRow()
}
