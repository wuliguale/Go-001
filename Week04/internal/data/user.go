package data

import "Go-001/Week04/internal/biz"

//do -> po

//实现biz层的接口
var _ biz.UserRepo = (*UserRepo)(nil)

type UserRepo struct {

}

func NewUserRepo() *UserRepo{
	return &UserRepo{}
}

func (userRepo *UserRepo) GetUser(uid int) (user *biz.User, err error){
	//todo
	return nil, nil
}


func (UserRepo *UserRepo) SaveUser(user *biz.User) bool {
	//todo
	return true
}

