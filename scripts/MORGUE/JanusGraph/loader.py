from rdf2g import setup_graph
import rdflib
import pathlib
from rdf2g import load_rdf2g


DEFAULT_LOCAL_CONNECTION_STRING = "ws://localhost:8182/gremlin"
g = setup_graph(DEFAULT_LOCAL_CONNECTION_STRING)

# Your data graph to load
OUTPUT_FILE_LAM_PROPERTIES = pathlib.Path("./data.nq").resolve()

rdf_graph = rdflib.Graph()
rdf_graph.parse(str(OUTPUT_FILE_LAM_PROPERTIES), format="nquads")

load_rdf2g(g, rdf_graph)
