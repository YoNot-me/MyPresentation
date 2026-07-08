package entity

type Config struct {
	Addr      string
	DBURL     string
	JWTKey    string
	TimeOut   string
	RedisPass string
}

type Admin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Brand struct {
	Name     string `json:"name" db:"name"`
	Password string `json:"password" db:"password"`
}

type Works struct {
	Brand       string `json:"brand" db:"brand"`
	WorkName    string `json:"work_name" db:"workName"`
	Url         string `json:"url" db:"url"`
	Description string `json:"description" db:"description"`
}

type Response struct {
	Data   any   `json:"data"`
	Status int   `json:"status"`
	Err    error `json:"err"`
}

type BrandsResponse struct {
	Name string `json:"name" db:"name"`
}

type WorksResponse struct {
	WorkName    string `json:"work_name" db:"workName"`
	Url         string `json:"url" db:"url"`
	Description string `json:"description" db:"description"`
}
