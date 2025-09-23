KNOWN ISSUES:
- 

TODO:
- create Dockerfile and get build_and_push working
- make resource Manager an interface so we can swap it out with non-memory solution
- Lightbits client implementation uses client + adapter, but we can actually just combine the two of them. we are separating them right now because we are reusing code from the old LB client implementation
- create cli interface for crud endpoints
    - create CLI command that will dump the internal mappings of resources to client. essentially, display a comprehensive view of StorMS
- Pinning service definition versions

THOUGHTS: 
- we want users to create StorMS to behave close to end user experience. that is, they will provide resource NAMES, since this is the customizable field. 
- should we strictly enforce having the patter than storms resource UUID is the vendor resource name?