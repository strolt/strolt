package api

import (
	"net/http"

	"github.com/strolt/strolt/apps/strolt/internal/config"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// type servicesResponseService struct {
// 	Services map[string]string `json:"services"`
// }

type servicesResponseTaskDestination struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
}

type servicesResponseTaskSource struct {
	Driver string `json:"driver"`
}

type servicesResponseTaskSchedule struct {
	Backup string `json:"backup"`
	Prune  string `json:"prune"`
}

type servicesResponseTask struct {
	Name          string                            `json:"name"`
	Source        servicesResponseTaskSource        `json:"source"`
	Notifications []string                          `json:"notifications"`
	Destinations  []servicesResponseTaskDestination `json:"destinations"`
	Schedule      servicesResponseTaskSchedule      `json:"schedule"`
}

type servicesResponseService struct {
	Tasks []servicesResponseTask `json:"tasks"`
}

type servicesResponse struct {
	Services map[string]servicesResponseService `json:"services"`
}

// ShowAccount godoc
// @Summary      Show an account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @success 200 {object} servicesResponse "desc"
// @Router       /metrics [get].
func servicesHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Compress(5)) //nolint:gomnd

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		c := config.Get()

		services := map[string]servicesResponseService{}
		for serviceName, service := range c.Services {
			tasks := []servicesResponseTask{}

			for taskName, task := range service {
				// fmt.Println("taskName", taskName)
				notifications := []string{}
				notifications = append(notifications, task.Notifications...)

				destinations := []servicesResponseTaskDestination{}

				for destinationName, destination := range task.Destinations {
					destinations = append(destinations, servicesResponseTaskDestination{
						Name:   destinationName,
						Driver: string(destination.Driver),
					})
				}

				tasks = append(tasks, servicesResponseTask{
					Name:          taskName,
					Notifications: notifications,
					Destinations:  destinations,
					Source: servicesResponseTaskSource{
						Driver: string(task.Source.Driver),
					},
					Schedule: servicesResponseTaskSchedule{
						Backup: task.Schedule.Backup,
						Prune:  task.Schedule.Prune,
					},
				})
			}

			services[serviceName] = servicesResponseService{
				Tasks: tasks,
			}
		}

		response := servicesResponse{
			Services: services,
		}

		render.JSON(w, r, response)
	})

	return r
}
