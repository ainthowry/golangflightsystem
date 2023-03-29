package cache

type UserRequest struct {
	ReqId  uint32
	Sender string
}

type ResponseData struct {
	Response []byte
}

// type ResponseCache struct {
// 	Cache map[UserRequest][]ResponseData
// }

// func NewRequestQueue() *ResponseCache {
// 	return &ResponseCache{
// 		Cache: make(map[UserRequest][]ResponseCache, 10),
// 	}
// }
