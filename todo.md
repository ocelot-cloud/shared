### TODO's

* "shared" needs a function like: logErrorAndRespond(w, status code, err) which logs the error and sends a json response with the error message
* problem: when doing assert.Nil, I only see where the assertion was called, not the stack trace -> add a deepstack log containing this information in the assert.Nil wrapper to get this information
* shared: input validator should create deepstack errors
* can "dev" branch be deleted? (local and remote)