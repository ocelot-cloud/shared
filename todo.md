### TODO's

* add deepstack log in all "shared" module errors
* "shared" needs a function like: logErrorAndRespond(w, status code, err) which logs the error and sends a json response with the error message -> also log level would be good I guess?
* problem: when doing assert.Nil, I only see where the assertion was called, not the stack trace -> add a deepstack log containing this information in the assert.Nil wrapper to get this information
* shared: input validator should create deepstack errors
* can "dev" branch be deleted? (local and remote)