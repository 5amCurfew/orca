function manualRun() {
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

function updateSidebar() {
    // Fetch the list of files in the "dags" directory
    // Replace this with your server-side logic to fetch the file list
    fetch('/dags')
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
    .then(response => response.json()) // Adjust based on your response format (JSON, HTML, etc.)
    .then(data => {
        d3.select("#graphPanel").selectAll("*").remove()
        console.log(data.graph)
        createDAGDiagram(data.graph)
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
    .then(response => response.json()) // Adjust based on your response format (JSON, HTML, etc.)
    .then(data => {
        // Update the HTML content in the graph panel
        panel.innerHTML = data.graph;
    })
    .catch(error => {
        console.error('Error fetching data:', error);
        // Handle errors as needed
    });
}

function createDAGDiagram(data) {
    const graphPanel = document.getElementById("graphPanel");

    const width = graphPanel.clientWidth;
    const height = 400;

    const svg = d3.select("#graphPanel").append("svg")
    .attr("width", width)
    .attr("height", height);

    const nodes = Object.keys(data.nodes).map(node => ({ id: node }));
    const links = [];

    for (const parent in data.children) {
        for (const child in data.children[parent]) {
            links.push({ source: nodes.find(n => n.id === parent), target: nodes.find(n => n.id === child) });
        }
    }

    const simulation = d3.forceSimulation(nodes)
    .force("charge", d3.forceManyBody().strength(-600))
    .force("link", d3.forceLink(links).strength(1).distance(100).iterations(10))
    .force("x", d3.forceX(width/2))
    .force("y", d3.forceY(height/2));

    const link = svg.selectAll(".link")
    .data(links)
    .enter().append("line")
    .attr("class", "link")
    .style( "stroke", "#000" );

    const node = svg.selectAll(".node")
    .data(nodes)
    .enter().append("circle")
    .attr("class", "node")
    .attr("r", 20);

    simulation.on("tick", () => {
    link
        .attr("x1", d => d.source.x)
        .attr("y1", d => d.source.y)
        .attr("x2", d => d.target.x)
        .attr("y2", d => d.target.y);

    node
        .attr("cx", d => d.x)
        .attr("cy", d => d.y);
    });
};