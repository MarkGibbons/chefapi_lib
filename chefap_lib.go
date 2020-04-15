
package chefapi_lib

// Auth is the structure returned to indicate access allowed or not
type Auth struct {
        Auth   bool   `json:"auth"`
        Group  string `json:"group"`
        Node   string `json:"node"`
        User   string `json:"user"`
}
