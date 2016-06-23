package snickers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/flavioribeiro/snickers/db"
	"github.com/flavioribeiro/snickers/rest"
	"github.com/flavioribeiro/snickers/types"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rest API", func() {
	Context("/jobs location", func() {
		var (
			response   *httptest.ResponseRecorder
			server     *mux.Router
			dbInstance db.DatabaseInterface
		)

		BeforeEach(func() {
			response = httptest.NewRecorder()
			server = rest.NewRouter()
			dbInstance, _ = db.GetDatabase()
			dbInstance.ClearDatabase()
		})

		It("GET should return application/json on its content type", func() {
			request, _ := http.NewRequest("GET", "/jobs", nil)
			server.ServeHTTP(response, request)
			Expect(response.HeaderMap["Content-Type"][0]).To(Equal("application/json; charset=UTF-8"))
		})

		It("GET should return stored jobs", func() {
			exampleJob1 := types.Job{ID: "123"}
			exampleJob2 := types.Job{ID: "321"}
			dbInstance.StoreJob(exampleJob1)
			dbInstance.StoreJob(exampleJob2)

			expected1, _ := json.Marshal(`[{"id":"123","source":"","destination":"","preset":{"video":{},"audio":{}},"status":"","progress":""},{"id":"321","source":"","destination":"","preset":{"video":{},"audio":{}},"status":"","progress":""}]`)
			expected2, _ := json.Marshal(`[{"id":"321","source":"","destination":"","preset":{"video":{},"audio":{}},"status":"","progress":""},{"id":"123","source":"","destination":"","preset":{"video":{},"audio":{}},"status":"","progress":""}]`)

			request, _ := http.NewRequest("GET", "/jobs", nil)
			server.ServeHTTP(response, request)
			responseBody, _ := json.Marshal(string(response.Body.String()))

			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(responseBody).To(SatisfyAny(Equal(expected1), Equal(expected2)))
		})

		It("POST should create a new job", func() {
			dbInstance.StorePreset(types.Preset{Name: "presetName"})
			jobJSON := []byte(`{"source": "http://flv.io/src.mp4", "destination": "s3://l@p:google.com", "preset": "presetName"}`)
			request, _ := http.NewRequest("POST", "/jobs", bytes.NewBuffer(jobJSON))
			server.ServeHTTP(response, request)

			jobs, _ := dbInstance.GetJobs()
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.HeaderMap["Content-Type"][0]).To(Equal("application/json; charset=UTF-8"))
			Expect(len(jobs)).To(Equal(1))
			job := jobs[0]
			Expect(job.Source).To(Equal("http://flv.io/src.mp4"))
			Expect(job.Destination).To(Equal("s3://l@p:google.com"))
			Expect(job.Preset.Name).To(Equal("presetName"))
		})

		It("POST should return BadRequest if preset is not set when creating a new job", func() {
			jobJSON := []byte(`{"source": "http://flv.io/src.mp4", "destination": "s3://l@p:google.com", "preset": "presetName"}`)
			request, _ := http.NewRequest("POST", "/jobs", bytes.NewBuffer(jobJSON))
			server.ServeHTTP(response, request)
			responseBody, _ := json.Marshal(string(response.Body.String()))

			expected, _ := json.Marshal(`{"error": "retrieving preset: preset not found"}`)
			Expect(responseBody).To(Equal(expected))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.HeaderMap["Content-Type"][0]).To(Equal("application/json; charset=UTF-8"))
		})

		It("GET to /jobs/jobID should return the job with details", func() {
			job := types.Job{
				ID:          "123-123-123",
				Source:      "http://source.here.mp4",
				Destination: "s3://ae@ae.com",
				Preset:      types.Preset{},
				Status:      "created",
				Progress:    "0%",
			}
			dbInstance.StoreJob(job)
			expected, _ := json.Marshal(&job)

			request, _ := http.NewRequest("GET", "/jobs/123-123-123", nil)
			server.ServeHTTP(response, request)
			Expect(response.Body.String()).To(Equal(string(expected)))
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.HeaderMap["Content-Type"][0]).To(Equal("application/json; charset=UTF-8"))
		})
	})
})