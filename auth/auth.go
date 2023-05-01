// Provides a simple shim for the BSD authentication API.
package auth

// #include <sys/types.h>
// #include <login_cap.h>
// #include <bsd_auth.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"os/user"
	"unsafe"
)

// Group to authenticate against, users in this group will have an
// "administrator" status in the application.
const authGroup = "operator"

type User struct {
	Name string
	Authenticated bool
}

func Authenticate(username string, password string) (User, error) {
	cUser := C.CString(username)
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cUser))
	defer C.free(unsafe.Pointer(cPassword))
	
	if err := UserCanAuthenticate(username); err != nil {
		return User{}, err
	}
	
	rc := C.auth_userokay(cUser, nil, nil, cPassword)
	if rc == 0 {
		return User{}, fmt.Errorf("quiltro/auth: authentication failed for %s",
			username)
	} else {
		return User{username, true}, nil
	}
}

func UserCanAuthenticate(username string) error {
	usr, err := user.Lookup(username)
	if err != nil {
		return err
	}

	groups, err := usr.GroupIds()
	if err != nil {
		return err
	}

	var groupFound bool
	for _, grp := range groups {
		/* I ignore the error because there's no way the OS will give us an
		   invalid ID. */
		group, _ := user.LookupGroupId(grp)
		if group.Name == authGroup {
			groupFound = true
			break
		}
	}

	if !groupFound {
		return fmt.Errorf("quiltro/auth: user %s is not in group %s",
			username, authGroup)
	} else {
		return nil
	}
}
