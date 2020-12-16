package service

import "Go-001/Week04/internal/biz"

//dto -> do

type UserService struct {
	UserBiz *biz.UserBiz
}


func NewUserService(userBiz *biz.UserBiz) *UserService{
	return &UserService{UserBiz:userBiz}
}

//返回grpc需要的对象
// 如果有错误，转换为http错误码
func (service *UserService) GetUserById(id int) () {
	//todo
}

//接收grpc传递的对象dto，转换为内部处理的对象do
//如果有错误，转换为http错误码
func (service *UserService) Register() {
	//todo
}

