## Config

Below is the current Gleaner config file example:


<div id="code-element"></div>
<script src="https://unpkg.com/axios/dist/axios.min.js"></script>
<script>
      axios({
      method: 'get',
      url: 'https://raw.githubusercontent.com/earthcubearchitecture-project418/gleaner/master/configs/template_v2.0.yaml'
       })
      .then(function (response) {
         document.getElementById("code-element").innerHTML = response.data;
      });
</script>
