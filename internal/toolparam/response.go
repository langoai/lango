package toolparam

// Response is a convenience alias for tool handler return values.
type Response map[string]interface{}

// StatusResponse creates a Response with a status field and optional extras.
func StatusResponse(status string, extras ...func(Response)) Response {
	r := Response{"status": status}
	for _, fn := range extras {
		fn(r)
	}
	return r
}

// ListResponse creates a Response with a named list and its count.
func ListResponse(key string, items interface{}, count int) Response {
	return Response{key: items, "count": count}
}
