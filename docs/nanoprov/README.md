# Nanopubs for PROV

## About 

Beginning to think about how Gleaner can represent the resources it harvests.
One thought is a nanopub on each resource.

How would I include something like:

* URL
* SHA256
* provided @ID (if provided)

Part of this grew out of the talk of "Terrior" in the ESIP Summer meeting.  It 
made me wonder if I could develop a finger print for the resource such that
it would make reconciling duplicated resources easier.  

While I like the idea that we should all have PIDs on all the things, I don't 
feel this is realistic.  I think we have to accept that there will be a lot of 
duplication across the net.  The goal will be to be able to know which ones are
likely the same.    If we provide a set of factors that are a "finger print" for 
a resource then we can likely present this back to the user.  


## Example


<div id="code-element"></div>
<script src="https://unpkg.com/axios/dist/axios.min.js"></script>
<script>
      axios({
      method: 'get',
      url: 'https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/docs/nanoprov/provNanoPub.jsonld'
       })
      .then(function (response) {
         document.getElementById("code-element").innerHTML = response.data;
      });
</script>

## References 

* http://nanoweb.dei.unipd.it/
* Indieweb specifications: https://indieweb.org/specifications
