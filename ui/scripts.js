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
        // Update the HTML content in the graph panel
        panel.innerHTML = data.graph;
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

// Play Button Click Event
const playButton = document.getElementById('runButton');
playButton.addEventListener('click', () => {
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
});