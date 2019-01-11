
package web

import "fmt"


type RequestError struct {
    Message string
    Code int
}

func (e *RequestError) Error() string {
    return fmt.Sprintf("%s", e.Message)
}
