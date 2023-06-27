package restliberror

type RestLibError struct {
	Err  error `json:"-"`
	Code int64 `json:"code"`
}

// Error Allows RestLibError to satisfy the error interface.
func (re RestLibError) Error() string {
	if re.Err == nil {
		return ""
	}
	return re.Err.Error()
}
