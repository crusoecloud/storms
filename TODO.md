KNOWN ISSUES:
- 

TODO:
- Lightbits client implementation uses client + adapter, but we can actually just combine the two of them. we are separating them right now because we are reusing code from the old LB client implementation
- Pinning service definition versions
- Consider enforcing usage of UUID instead of string. this will ultimately be more limiting but can offer benefits with checking args
- Design an o11y cli subcommand that will show useful cluster/resource information
    - create CLI command that will dump the internal mappings of resources to client. essentially, display a comprehensive view of StorMS
- need to configure a HV host to be able to talk to multiple lightbits backends 
    - for mvp, we can do this manually, but we need a solution at least 
    - later, we can have it talk to multiple vendors
- make storms respond gracefully to if any downstream clients are down. we dont want a broken backend to crash storms 
- evaluate the difference between `storms sync --all` and `storms app reload`
- remove reference from the internal tools repo to support open-source

THOUGHTS: 
- should we strictly enforce having the pattern than storms resource UUID is the vendor resource name?