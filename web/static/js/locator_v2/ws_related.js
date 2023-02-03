var host = document.location.host;
var path = document.location.pathname;
var timecounter = setInterval(function () {
}, 1000000)
var conn
var message

// Das hier weils wichtig ist ob https oder nicht.
// Browser erlauben KEIN downgrad also https zu ws!
if (location.protocol === 'https:') {
    conn = new WebSocket("wss://" + document.location.host + path + "/ws?user=" + user)
} else {
    conn = new WebSocket("ws://" + document.location.host + path + "/ws?user=" + user)
}

// console.log(conn)

conn.onopen = function (evt) {
    // console.log("Success", conn, evt)
};
conn.onclose = function (evt) {
    // console.log("Closed", conn, evt)
};
conn.onmessage = function (evt) {
    // console.log("Message", evt.data)
    // message = JSON.parse(JSON.stringify(evt.data))
    message = JSON.parse(evt.data)
    console.log(message)


    if (message.status !== "psub") {
        // document.getElementById("state").innerHTML = message.state
        document.getElementById("countdown").innerHTML = message.time
        clearInterval(timecounter)
        timecounter = setInterval(function () {
            document.getElementById("countdown").innerHTML -= 1
        }, 1000);
    }

    if (message.status === "location") {
        // document.getElementById("city").innerHTML = message.Location
        document.getElementById("locationteller").innerHTML = message.Location
        document.getElementById("status").innerHTML = message.status
        document.getElementById("rounds").innerHTML = message.rounds - 1
        document.getElementById("submitButton").disabled = false;
        // document.getElementById("points").innerHTML = ""
        document.getElementById("distance").innerHTML = ""
        document.getElementById("awarded").innerHTML = ""
        solution_layer.getSource().clear()
        submit_Layer.getSource().clear()
        removeSolutionPopup()

    }
    if (message.status === "psub") {
        // document.getElementById("psub").innerHTML = message.Player + " submitted"

        // Appraoch is good but position does not work like this...
        // var x = map.getView().calculateExtent()
        // minx = x[0]
        // miny = x[1]
        // maxx = x[2]
        // maxy = x[3]

        // notifierOL.setPosition([maxx - maxx*0.5, maxy - maxy * 0.2])
        notifierOL.setPosition(map.getView().getCenter())
        notifierContent.innerHTML = message.Player + " submitted"
        x = setTimeout(function () {
            // document.getElementById("psub").innerHTML = ""
            notifierOL.setPosition(undefined)
        }, 1000)
    }
    if (message.status === "reviewing") {
        document.getElementById("status").innerHTML = message.status
        document.getElementById("points").innerHTML = JSON.stringify(message.points)
        document.getElementById("distance").innerHTML = (message.distance / 1000).toFixed(2) + " km away. "
        document.getElementById("awarded").innerHTML = message.awarded + " Points"
        document.getElementById("submitButton").disabled = true;
        addSolution(message)
        addCommit(message)
        displaySolutionPopup(message)


    }
    if (message.status === "finished") {
        clearInterval(timecounter)
        document.getElementById("status").innerHTML = message.status
        document.getElementById("points").innerHTML = JSON.stringify(message.points)
        document.getElementById("distance").innerHTML = ""
        document.getElementById("awarded").innerHTML = ""
        document.getElementById("countdown").innerHTML = ""
        document.getElementById("locationteller").innerHTML = "Finished"
        solution_layer.getSource().clear()
        removeSolutionPopup()
        conn.close()
    }
};

function submitGuess() {
    coords = point.getGeometry().getCoordinates()
    coords4326 = ol.proj.transform(coords, 'EPSG:3857', 'EPSG:4326')
    // console.log({"lat":coords4326[0], "long": coords4326[1]})

    x = {"lat": coords4326[1], "long": coords4326[0]}
    conn.send(JSON.stringify(x));
}

document.getElementById("submitButton").onclick = function () {
    submitGuess();
    // console.log("Send ready message!")
};

console.log("ws_related loaded successfully")

function addSolution(message) {
    if (message.geom === "Point") {
        solution = new ol.Feature({
            geometry: new ol.geom.Point(ol.proj.transform([message.lng, message.lat], 'EPSG:4326', 'EPSG:3857'))
        });
        solution_layer.getSource().addFeature(solution)
        // flyTo(ol.proj.fromLonLat([message.lng, message.lat]), function () {
        // });
        console.log(solution)
        return
    }
    solution = new ol.format.GeoJSON().readFeatures(message.geojson);
    solution_layer.getSource().addFeatures(solution)
    // flyTo(ol.proj.fromLonLat([message.lng, message.lat]), function () {
    // });
}

function addCommit(message) {
    for (const [key, value] of Object.entries(message.submits)) {
        // console.log(`${key}: ${value}`);

        if (value[0] === 0) {
            continue
        }

        solution = new ol.Feature({
            name: key,
            geometry: new ol.geom.Point(ol.proj.transform([value[1], value[0]], 'EPSG:4326', 'EPSG:3857'))
        });
        submit_Layer.getSource().addFeature(solution)
    }

}

function displaySolutionPopup(message) {
    const coordinate = ol.proj.transform([message.lng, message.lat], 'EPSG:4326', 'EPSG:3857')

    popup.setPosition(coordinate);
}

function removeSolutionPopup() {
    popup.setPosition(undefined);
    closer.blur();
    return false;
}