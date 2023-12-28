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

    // Get the dimensions of the parent div
    const width = graphPanel.clientWidth;
    const height = 600;

    const svg = d3.select("#graphPanel").append("svg")
        .attr("width", width)
        .attr("height", height);

    // Add arrowhead marker definition
    svg.append("defs").append("marker")
        .attr("id", "arrowhead")
        .attr("viewBox", "0 -5 10 10")
        .attr("refX", 80)
        .attr("refY", 0)
        .attr("markerWidth", 20)
        .attr("markerHeight", 10)
        .attr("orient", "auto")
        .append("path")
        .attr("d", "M0,-5L10,0L0,5")
        .attr("class", "arrowhead-path");

    const nodes = Object.keys(data.nodes).map(node => ({ id: node }));
    const links = [];

    for (const parent in data.children) {
        for (const child of data.children[parent]) {
            links.push({ source: nodes.find(n => n.id === parent), target: nodes.find(n => n.id === child) });
        }
    }

    const simulation = d3.forceSimulation(nodes)
        .force("charge", d3.forceManyBody().strength(-200))
        .force("center", d3.forceCenter(width / 2, height / 2))
        .force("link", d3.forceLink(links).id(d => d.id).strength(0.1).distance(150).iterations(10));

    const link = svg.selectAll(".link")
        .data(links)
        .enter().append("line")
        .attr("class", "link")
        .attr("marker-end", "url(#arrowhead)")  // Add arrowhead marker
        .style("stroke", "#000");

    const node = svg.selectAll(".node")
        .data(nodes)
        .enter().append("rect") // Use "rect" instead of "circle" for rectangles
        .attr("class", "node")
        .attr("width", d => (d.id.length * 10) + 20) // Set the width of the rectangle
        .attr("height", 20) // Set the height of the rectangle
        .attr("stroke", "black")
        .attr("fill", "white");

    const labels = svg.selectAll(".label")
        .data(nodes)
        .enter().append("text")
        .attr("class", "label")
        .text(d => d.id)
        .attr("text-anchor", "middle") // Center the text horizontally
        .attr("dominant-baseline", "middle") // Center the text vertically

    simulation.on("tick", () => {
        link
            .attr("x1", d => d.source.x)
            .attr("y1", d => d.source.y)
            .attr("x2", d => d.target.x)
            .attr("y2", d => d.target.y);

        node
            .attr("x", d => d.x - (d.id.length * 10 + 20)/2) // Adjust the positioning based on rectangle width
            .attr("y", d => d.y); // Adjust the positioning based on rectangle height

        labels
            .attr("x", d => d.x) // Adjust based on half the width of the rectangle
            .attr("y", d => d.y + 10); // Adjust based on half the height of the rectangle
    });
}