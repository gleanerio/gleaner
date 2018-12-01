### Framing

## Notes
So this is a special section for the crawler code in ways.  The framing is more a "service" 
than the others.  For example the citation and organizational framing are NOT required for 
any index here.  They might be useful to others or to a particular view on a record.  Something
more like what a front end UI might do (though that UI could process the frame locally
and not need a RPC).  

The spatial frame is the only one that is useful as it extracts a particular frame/view that
aids in the actual indexing of that resource.  

So, these are here for now, and some code in here might be useful.  However, it's likely
these should be a GRPC service (or FN?) that is called by cralwer, or webui or a component via 
service end point as needed.   Not as a canonical part of this application.

