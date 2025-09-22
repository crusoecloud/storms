KNOWN ISSUES:
- 

TODO:
- support lightbits snapshots
- create Dockerfile and get build_and_push working
- make resource Manager an interface so we can swap it out with non-memory solution
- implement ClientAllocation algo/function for testability and ease to transition to lowest-usage allocation
    - where do we specify what allocation algo to use?
    - if we should, how should we separate app and service configuration?
- Lightbits client implementation uses client + adapter, but we can actually just combine the two of them. we are separating them right now because we are reusing code from the old LB client implementation
