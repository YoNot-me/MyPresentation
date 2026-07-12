package server

import (
	"net/http"
	"presentator/internal/features/auth/token"
	"time"

	"presentator/internal/core/entity"
	"presentator/internal/core/middleware"
	adminTransport "presentator/internal/features/admin/transport"
	authTransport "presentator/internal/features/auth/transport"
	fileservingTransport "presentator/internal/features/fileserving/transport"
)

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Server struct {
	Http   *http.Server
	DB     *pgxpool.Pool
	Logger *zap.Logger
}

func ServerInit(
	env *entity.Config,
	db *pgxpool.Pool,
	openLog *zap.Logger,
	fileServing *fileservingTransport.FileServingTransport,
	auth *authTransport.AuthTransport,
	admin *adminTransport.AdminTransport,
	jwt *JWT.ServingJWT,
) (*Server, error) {

	mux := route(fileServing, auth, admin, jwt)

	Server := &Server{
		Http: &http.Server{
			Addr: env.Addr,
			Handler: http.TimeoutHandler(
				mux,
				180*time.Second,
				`{"error":"request timeout"}'`,
			),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 180 * time.Second,
		},
		DB:     db,
		Logger: openLog,
	}

	return Server, nil
}

func route(

	fileServing *fileservingTransport.FileServingTransport,
	auth *authTransport.AuthTransport,
	admin *adminTransport.AdminTransport,
	jwt *JWT.ServingJWT,
) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	//middleware
	r.Use(gin.Recovery())
	r.Use(middleware.MaximumSize())

	//basic
	r.Handle(http.MethodGet, "/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/works")
	})

	//auth
	r.Handle(http.MethodGet, "/auth", auth.Auth)
	r.Static("/auth", "./public/auth")

	r.Handle(http.MethodPost, "/auth/check", auth.AuthBrand)
	r.Handle(http.MethodPost, "/logout", auth.Logout)

	r.Handle(http.MethodGet, "/admin", admin.MainAdmin)
	r.Handle(http.MethodPost, "/admin/auth", admin.AuthAdmin)
	r.Handle(http.MethodPost, "/logout/admin", admin.LogOut)

	//fileserving
	protect := r.Group("")
	protect.Use(middleware.Protected(jwt))
	protect.Handle(http.MethodGet, "/works", fileServing.ListWorkFiles)
	protect.Handle(http.MethodGet, "/works/files/:name", fileServing.ListWorkImages)
	protect.Handle(http.MethodGet, "/presentation/:name/*filepath", fileServing.GetWork)

	protect.Handle(http.MethodGet, "/works/serve", fileServing.ServeHTML)

	r.Static("/static/presentation", "./public/presentation")

	//admin
	protectAdmin := r.Group("/admin")
	protectAdmin.Use(middleware.ProtectedAdmin(jwt))

	protectAdmin.Handle(http.MethodGet, "/brands", admin.ListBrands)
	protectAdmin.Handle(http.MethodPost, "/brands/add", admin.AddNewBrand)
	protectAdmin.Handle(http.MethodPut, "/:brandName/rename", admin.RenameBrand)
	protectAdmin.Handle(http.MethodDelete, "/:brandName", admin.DeleteBrand)
	protectAdmin.Handle(http.MethodPut, "/:brandName/password", admin.ChangeBrandPassword)

	protectAdmin.Handle(http.MethodGet, "/:brandName/works", admin.ListAllBrandWorks)
	protectAdmin.Handle(http.MethodPost, "/:brandName/works/add", admin.AddNewWork)
	protectAdmin.Handle(http.MethodDelete, "/:brandName/remove/:workName", admin.DeleteWork)
	protectAdmin.Handle(http.MethodPut, "/:brandName/:workName/change", admin.ChangeWorkFields)

	protectAdmin.Handle(http.MethodGet, "/:brandName/files/:workName", admin.ListWorkImages)
	protectAdmin.Handle(http.MethodGet, "/:brandName/serve/:workName/*filepath", admin.ServingWork)
	protectAdmin.Static("/panel", "./public/admin")

	return r
}

func (s *Server) Run() error {
	if err := s.Http.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
