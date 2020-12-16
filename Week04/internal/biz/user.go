package biz

type UserBiz struct {
	UserRepo *UserRepo
}


type User struct {
	Id int
	Name string
	Age int
	Password string
	//内部传递的对象do，添加此属性表示与dto和po都不相同
	Rand int
}


//定义data层的接口
type UserRepo interface {
	GetUser(int) (*User, error)
	SaveUser(*User) bool
}


func NewUserBiz(userRepo *UserRepo) *UserBiz {
	return &UserBiz{UserRepo:userRepo}
}


func (userBiz *UserBiz) InsertOrUpdate(user *User) (*User, error) {
	//todo
	userBiz.UserRepo.GetUser(1)

	return nil, nil
}

