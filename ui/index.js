// ////////////////////////////////////
// Pulse
// ////////////////////////////////////
function refresh() {
    // Fetch the list of files in the "dags" directory
    fetch('/refresh')
        .then(response => response.json())
        .then(data => {
            const dagButtonsContainer = document.getElementById('dagButtons');
            dagButtonsContainer.innerHTML = '';
            // Create a button for each file
            data.dagList.forEach(fileName => {
                const button = document.createElement('button');
                button.textContent = fileName;
                button.addEventListener('click', () => {
                    window.location.hash = fileName;
                });
                dagButtonsContainer.appendChild(button);
            });
        });
}


// ////////////////////////////////////
// Execute
// ////////////////////////////////////
function execute() {
    const currentHash = window.location.hash.substring(1);

    // Send a POST request to the execute route
    fetch('/execute', {
        method: 'POST',
        body: JSON.stringify({ file_path: 'dags/'+currentHash }),
    })
    .then(response => response.json())
    .then(data => {
        // Handle the response if needed
        console.log('Execution response:', data);
    })
    .catch(error => {
        console.error('Error executing DAG:', error);
    });
};


// ////////////////////////////////////
// Update
// ////////////////////////////////////
function updateGraphPanel() {
    const currentHash = window.location.hash.substring(1);

    fetch(`/graph`, {
        method: "POST",
        body: JSON.stringify({ file_path: 'dags/'+currentHash }),
    })
    .then(response => response.json())
    .then(data => {
        // console.log(data);
        d3.select("#graphPanelSVG").select("g").selectAll(".output").remove()
        createTreeDiagram(data.graph);
    })
    .catch(error => {
        console.error('Error fetching data:', error);
        // Handle errors as needed
    });
}

function updateStatusPanel() {
    const currentHash = window.location.hash.substring(1);

    fetch(`/status`, {
        method: "POST",
        body: JSON.stringify({ file_path: 'dags/'+currentHash }),
    })
    .then(response => response.json())
    .then(data => {
        //console.log(data);
    })
    .catch(error => {
        console.error('Error fetching data:', error);
        // Handle errors as needed
    });
}


// ////////////////////////////////////
// Create Tree Diagram
// ////////////////////////////////////
function transformRepresentation(graph) {
    const nodes = [];
    const edges = [];
  
    // Extract nodes
    Object.keys(graph.tasks).forEach(taskKey => {
      const task = graph.tasks[taskKey];
      nodes.push({ id: task.name, desc: task.desc });
    });
  
    // Extract edges
    Object.keys(graph.children).forEach(parentKey => {
      const parent = graph.children[parentKey];
      Object.keys(parent).forEach(childKey => {
        edges.push({ source: parentKey, target: childKey });
      });
    });
    
    return [nodes, edges];
};

function createTreeDiagram(data) {
    var g = new dagreD3.graphlib.Graph().setGraph({
        nodesep: 50,
        ranksep: 50,
        rankdir: "LR",
        marginx: 10,
        marginy: 10
      });

    // Data for this example
    var [nodes, edges] = transformRepresentation(data)

    // Add nodes
    nodes.forEach(node => {  
        g.setNode(node.id, {
            labelType: "html",
            label: `<div class='nodeName'>${node.id}</div>`,
            width: node.id.length * 10,
            rx: 5,
            ry: 5
        }); 
    });  

    // Add edges
    edges.forEach(edge => {
        g.setEdge(edge.source, edge.target, {
            arrowhead: "normal",
            arrowheadStyle: "fill: #383838",
            lineInterpolate: 'basis',
            curve: d3.curveBasis
        });
    });

    var render = new dagreD3.render();

    var svg = d3.select("#graphPanelSVG"),
    inner = svg.select("g"),
    zoom = d3.zoom().on("zoom", function() {
      inner.attr("transform", d3.zoomTransform(this));
    });
    svg.call(zoom);
  
    inner.call(render, g);
  };
  