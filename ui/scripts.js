function Execute() {
    const currentHash = window.location.hash.substring(1); // Exclude the '#' symbol

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

function pulse() {
    // Fetch the list of files in the "dags" directory
    fetch('/pulse')
        .then(response => response.json())
        .then(data => {
            const dagButtonsContainer = document.getElementById('dagButtons');
            dagButtonsContainer.innerHTML = '';
            // Create a button for each file
            data.forEach(fileName => {
                const button = document.createElement('button');
                button.textContent = fileName;
                button.addEventListener('click', () => {
                    window.location.hash = fileName;
                });
                dagButtonsContainer.appendChild(button);
            });
        });
}

function updateGraphPanel() {
    const panel = document.getElementById('graphPanel');
    const currentHash = window.location.hash.substring(1); // Exclude the '#' symbol

    fetch(`/graph`, {
        method: "POST",
        body: JSON.stringify({ file_path: 'dags/'+currentHash }),
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
        d3.select("#graphPanel").selectAll("*").remove()
        createTreeDiagram(data.graph.tasks);
    })
    .catch(error => {
        console.error('Error fetching data:', error);
        // Handle errors as needed
    });
}

function updateStatusPanel() {
    const panel = document.getElementById('statusPanel');
    const currentHash = window.location.hash.substring(1); // Exclude the '#' symbol

    fetch(`/status`, {
        method: "POST",
        body: JSON.stringify({ file_path: 'dags/'+currentHash }),
    })
    .then(response => response.json())
    .then(data => {
        //console.log('updateStatusPanel')
    })
    .catch(error => {
        console.error('Error fetching data:', error);
        // Handle errors as needed
    });
}

function createTreeDiagram(data) {

  const result = [];

  for (const taskName in data) {
    const task = data[taskName];
    const modifiedTask = {
      name: task.name,
      desc: task.desc,
      cmd: task.cmd,
      status: task.status,
    };

    if (task.children) {
      modifiedTask.children = task.children.map((child) => ({
        name: child.name,
        desc: child.desc,
        cmd: child.cmd,
        status: child.status,
      }));
    }

    result.push(modifiedTask);
  }

  var treeData = {
    "name": "root",
    "children": result
  }

  // set the dimensions and margins of the diagram
  const margin = {top: 20, right: 90, bottom: 30, left: 90},
        width  = 600 - margin.left - margin.right,
        height = 600 - margin.top - margin.bottom;
  
  // declares a tree layout and assigns the size
  const treemap = d3.tree().size([height, width]);
  
  //  assigns the data to a hierarchy using parent-child relationships
  let nodes = d3.hierarchy(treeData, d => d.children);
  // maps the node data to the tree layout
  nodes = treemap(nodes);
  
  // append the svg object to the body of the page
  // appends a 'group' element to 'svg'
  // moves the 'group' element to the top left margin
  const svg = d3.select("#graphPanel").append("svg")
          .attr("width", width + margin.left + margin.right)
          .attr("height", height + margin.top + margin.bottom),
        g = svg.append("g")
          .attr("transform",
              "translate(" + margin.left + "," + margin.top + ")");
  
  // adds the links between the nodes
  const link = g.selectAll(".link")
    .data( nodes.descendants().slice(1))
    .enter().append("path")
    .attr("class", "link")
    .style("stroke", "blue")
    .attr("fill", "none")
    .attr("d", d => {
        return "M" + d.y + "," + d.x
        + "C" + (d.y + d.parent.y) / 2 + "," + d.x
        + " " + (d.y + d.parent.y) / 2 + "," + d.parent.x
        + " " + d.parent.y + "," + d.parent.x;
        });
  
  // adds each node as a group
  const node = g.selectAll(".node")
      .data(nodes.descendants())
      .enter().append("g")
      .attr("transform", d => "translate(" + d.y + "," + d.x + ")");
  
  // adds the circle to the node
  node.append("circle")
    .attr("r", d => 10)
    .style("stroke", "black")
    .style("fill", "white");
    
  // adds the text to the node
  node.append("text")
    .attr("dy", ".35em")
    .attr("x", d => d.children ? (10 + 5) * -1 : 10 + 5)
    .attr("y", d => d.children && d.depth !== 0 ? -(10 + 5) : 10)
    .style("text-anchor", d => d.children ? "end" : "start")
    .text(d => d.data.name);
};
