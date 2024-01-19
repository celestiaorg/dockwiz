package api

const endpointPrefix = "/api/v1"

type serviceEndpointPath struct{}

var APIPath = &serviceEndpointPath{}

func (e *serviceEndpointPath) Status() string {
	return endpointPrefix + "/status/{image_id}"
}

func (e *serviceEndpointPath) Build() string {
	return endpointPrefix + "/build"
}
