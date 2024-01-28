// ////////////////////////////////////
// Pulse
// ////////////////////////////////////
function refresh() {
    const dagButtonsContainer = document.getElementById('dagButtons');

    fetch('/refresh')
        .then(response => response.json())
        .then(data => {
            dagButtonsContainer.innerHTML = '';
            // Create a button for each file
            data.dagList.forEach(fileName => {
                const button = document.createElement('button');
                button.textContent = fileName;
                button.addEventListener('click', () => {
                    window.location.hash = fileName.split(".orca")[0];
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
    const dag = currentHash.split("@")[0];

    // Send a POST request to the execute route
    fetch('/execute', {
        method: 'POST',
        body: JSON.stringify({ path: `dags/${dag}` }),
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
    const dag = currentHash.split("@")[0];
    const executionStart = currentHash.split("@")[1];

    if(dag){
        fetch(`/graph`, {
            method: "POST",
            body: JSON.stringify({ path: `dags/${dag}` }),
        })
        .then(response => response.json())
        .then(data => {
            d3.select("#graphPanelSVG").select("g").selectAll(".output").remove()
            createTreeDiagram(data.graph);
        })
        .catch(error => {
            console.error('Error fetching data:', error);
            // Handle errors as needed
        });
    }
}

function updateLogsPanel() {
    const currentHash = window.location.hash.substring(1);
    const dag = currentHash.split("@")[0];
    const executionStart = currentHash.split("@")[1];

    const logButtonsContainer = document.getElementById('logButtons');

    if(dag){
        fetch(`/executionLogs`, {
            method: "POST",
            body: JSON.stringify({ path: `logs/${dag}`}),
        })
        .then(response => response.json())
        .then(data => {
            logButtonsContainer.innerHTML = '';
            n = data.logList.length < 10? data.logList.length : 10;
            data.logList.slice(-n).reverse().forEach(dirName => {
                const button = document.createElement('button');
                
                button.textContent = dirName;
                logButtonsContainer.appendChild(button);
                button.addEventListener('click', () => {
                    const d = window.location.hash.split("@")[0];
                    window.location.hash = `${d}@${dirName}`;
                });
            });
        })
        .catch(error => {
            console.error('Error fetching data:', error);
            // Handle errors as needed
        });
    }
}

function updateLogTasksPanel() {
    const currentHash = window.location.hash.substring(1);
    const dag = currentHash.split("@")[0];
    const executionStart = currentHash.split("@")[1];

    const logButtonsContainer = document.getElementById('logTaskButtons');

    if(dag && executionStart){
        fetch(`/executionTaskLogs`, {
            method: "POST",
            body: JSON.stringify({ path: `logs/${dag}/${executionStart}` }),
        })
        .then(response => response.json())
        .then(data => {
            logButtonsContainer.innerHTML = '';
            data.logTaskList.forEach(fileName => {
                const button = document.createElement('button');
                button.textContent = fileName;
                button.addEventListener('click', () => {
                    const d = `${window.location.hash.split("@")[0]}@${window.location.hash.split("@")[1]}`;
                    window.location.hash = `${d}@${fileName}`;
                });
                logButtonsContainer.appendChild(button);
            });
        })
        .catch(error => {
            //console.error('Error fetching data:', error);
            // Handle errors as needed
        });
    } else{
        logButtonsContainer.textContent = '';
    }
}

function updateLogTaskViewer() {
    const currentHash = window.location.hash.substring(1);
    const dag = currentHash.split("@")[0];
    const executionStart = currentHash.split("@")[1];
    const task = currentHash.split("@")[2];

    const logTaskViewerOutput = document.getElementById('logTaskViewerOutput');

    if(dag && executionStart && task){
        fetch(`/executionTaskLog`, {
            method: "POST",
            body: JSON.stringify({ path: `logs/${dag}/${executionStart}/${task}` }),
        })
        .then(response => response.json())
        .then(data => {
            logTaskViewerOutput.textContent = data.log;
        })
    } else{
        logTaskViewerOutput.textContent = '';
    }
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
  