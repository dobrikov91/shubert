let socket;

function loadInitialData() {
    fetch('/init')
        .then(response => response.json())
        .then(data => {
            updatePage(data);
            connectWebSocket(); // Start WebSocket connection after loading initial data
        })
        .catch(error => {
            console.error("Error loading initial data:", error);
        });
}

function highlightElement(data) {
    id = data.highlightId - 1
    const element = document.getElementById(`Command${id}`);

    if (element) {
        // element.scrollIntoView({ behavior: "smooth"} );
        const y = element.getBoundingClientRect().top + window.scrollY - window.innerHeight / 3;
        window.scrollTo({top: y, behavior: 'smooth'});

        // const style = getComputedStyle(element);

        if (!element.classList.contains("highlight")) {
            setTimeout(() => {
                element.classList.remove("highlight");
                element.style.setProperty("box-shadow", "0 0 0px #00ff73");
            }, 500);
        }

        element.classList.add("highlight");
        element.style.setProperty("box-shadow", "0 0 20px #00ff73");
    }
}

function escapeHtml(str) {
    return str
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

function delIndex(index) {
    url = `/delete?index=${index}`
    var xhr = new XMLHttpRequest();
    xhr.open("POST", url, true);
    xhr.send();
}

function updatePage(data) {
    const container = document.getElementById(`card-holder`);
    container.innerHTML = '';

    var devices = []
    data.commands.forEach((cmd, index) => {
        if (devices.indexOf(cmd.event.device) == -1) {
            devices.push(cmd.event.device)
        }
    })
    devices.sort()

    devices.forEach((device) => {
        const deviceDetails = document.createElement('details')
        deviceDetails.className = 'device-name'
        deviceDetails.id = `DeviceDetails${device}`
        deviceDetails.innerHTML = `<summary>Device ${device}</summary>`
        container.appendChild(deviceDetails);

        const id = deviceDetails.id
        const isOpen = localStorage.getItem(`details-${id}`) === "true";
        deviceDetails.open = isOpen;

        // Save state when toggled
        deviceDetails.addEventListener("toggle", () => {
            localStorage.setItem(`details-${id}`, deviceDetails.open);
        });

        const deviceContainer = document.createElement('container')
        deviceContainer.className = 'card-holder'
        deviceContainer.id = `Device${device}`
        deviceDetails.appendChild(deviceContainer)
    })

    data.commands.forEach((cmd, index) => {
        const fieldset = document.createElement('fieldset');
        fieldset.className = 'card';
        fieldset.id = `Command${index}`;

        fieldset.innerHTML = `
        <div class="form-row">
            <div class="card-button-holder">
                <button class="card-button" id="card-button-yes">✓</button>
                <button type="button" data-index=${index} formaction=javascript:delIndex(${index}) formmethod="post" class="card-button delete-button" id="card-button-no">╳</button>
            </div>

            <input class="card-alias" type="text" name="Alias" placeholder="Alias" value="${cmd.alias}"></input>

            <input id="Device" name="Device" type="hidden" value="${cmd.event.device}" readonly>
            <input id="Channel" name="Channel" type="hidden" type="text" value="${cmd.event.channel}" readonly>
            <input id="Key" name="Key" type="hidden" type="text" value="${cmd.event.key}" readonly>

            <textarea class="card-command" name="Command" placeholder="Command">${escapeHtml(cmd.command)}</textarea>

            <div class="card-params-holder">
                <label class="card-params-label" for="Trigger${index}">Trigger</label>
                <select class="card-params-select" id="Trigger${index}" name="Trigger">
                    <option value="OnPress" ${cmd.trigger == 'OnPress' ? 'selected' : ''}>OnPress</option>
                    <option value="OnRelease" ${cmd.trigger == 'OnRelease' ? 'selected' : ''}>OnRelease</option>
                    <option value="OnChange" ${cmd.trigger == 'OnChange' ? 'selected' : ''}>OnChange</option>
                </select>

                <label class="card-params-label" for="Timeout${index}">Timeout(ms)</label>
                <input class="card-params-number" id="Timeout${index}" type="number" name="Timeout" placeholder="0" min="0" value="${cmd.timeout_ms}"></input>
            </div>
        </div>
        
        <div id="confirm-modal${index}" class="modal">
            <div class="modal-content">
                <p>Are you sure you want to delete this item?</p>
                <div class="modal-buttons">
                    <button type="button" class="modal-button-yes" id="confirm-delete${index}">Yes</button>
                    <button type="button" class="modal-button-no" id="cancel-delete${index}">Cancel</button>
                </div>
            </div>
        </div>
        `;
        const element = document.getElementById(`Device${cmd.event.device}`);
        element.appendChild(fieldset)
    });

    let idToDelete = null;

    document.querySelectorAll(".delete-button").forEach(btn => {
        const modal = document.getElementById(`confirm-modal${btn.dataset.index}`);
        const confirmBtn = document.getElementById(`confirm-delete${btn.dataset.index}`);
        const cancelBtn = document.getElementById(`cancel-delete${btn.dataset.index}`);

        btn.addEventListener("click", function (e) {
            e.preventDefault();
            idToDelete = this.dataset.index
            modal.style.display = "flex";
        });

        confirmBtn.addEventListener("click", function () {
            modal.style.display = "none";
            if (idToDelete) {
                delIndex(idToDelete);
                idToDelete = null;
            }
        });

        cancelBtn.addEventListener("click", function () {
            idToDelete = null;
            modal.style.display = "none";
        });
    });
}

// Connect WebSocket on page load
document.addEventListener('DOMContentLoaded', loadInitialData);


editMode = document.getElementById('mode')
editMode.addEventListener('change', function(){
    var xmlHttp = new XMLHttpRequest();
    button = document.getElementById("mapping-button")
    if (this.checked) {
        xmlHttp.open("GET", "/modeEdit", true );
        xmlHttp.send( null );
        button.innerText = "Stop mapping";
    } else {
        xmlHttp.open("GET", "/modeCommands", true );
        xmlHttp.send( null );
        button.innerText = "Start mapping";
    }
})
