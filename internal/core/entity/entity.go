package entity

type Config struct {
	Addr    string
	DBURL   string
	JWTKey  string
	TimeOut string
}

type Admin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// сущности бренда и файлов
type Brand struct {
	Name     string `json:"name" db:"name"`
	Password string `json:"password" db:"password"`
}

type Works struct {
	Id       int    `json:"id" db:"id"`
	Brand    string `json:"brand" db:"brand"`
	WorkName string `json:"work_name" db:"workName"`
	Url      string `json:"url" db:"url"`
}

type Response struct {
	Data   any   `json:"data"`
	Status int   `json:"status"`
	Err    error `json:"err"`
}
