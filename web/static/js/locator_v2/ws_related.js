var list = document.getElementById('demo');   
var host = document.location.host;
var path = document.location.pathname;
var timecounter =setInterval(function(){},1000000)
var conn

// Das hier weils wichtig ist ob https oder nicht.
// Browser erlauben KEIN downgrad also https zu ws!
if (location.protocol === 'https:'){
    conn = new WebSocket("wss://" + document.location.host + path + "/ws?user=" + user)
} else {
    conn = new WebSocket("ws://" + document.location.host + path + "/ws?user=" + user)
};

// console.log(conn)

conn.onopen = function (evt) {
    // console.log("Success", conn, evt)
};
conn.onclose = function (evt) {
    // console.log("Closed", conn, evt)
};
conn.onmessage = function (evt) {
    console.log("Message", evt.data)
    //message = JSON.parse(JSON.stringify(evt.data))
    message = JSON.parse(evt.data)
    console.log(message)

    // document.getElementById("state").innerHTML = message.state
    document.getElementById("time").innerHTML = message.time
    //timecounter = doTimestuff(message.time)

    clearInterval(timecounter)
    timecounter = setInterval(function () {document.getElementById("time").innerHTML -= 1}, 1000); 

    if (message.status == "location") {
        document.getElementById("city").innerHTML = message.Location
        document.getElementById("status").innerHTML = message.status
        document.getElementById("points").innerHTML = ""
        document.getElementById("distance").innerHTML =""
        document.getElementById("awarded").innerHTML =""
        solution_layer.getSource().clear()
        
    }
    if (message.status == "reviewing") {
        document.getElementById("status").innerHTML = message.status
        document.getElementById("points").innerHTML = JSON.stringify(message.points)
        document.getElementById("distance").innerHTML = message.distance +  " km away. "
        document.getElementById("awarded").innerHTML = "--> "+ message.awarded + " Points"
        addSolution(message)
        }
};

function submitGuess(){
    coords = point.getGeometry().getCoordinates()
    coords4326 = ol.proj.transform(coords, 'EPSG:3857', 'EPSG:4326')
    // console.log({"lat":coords4326[0], "long": coords4326[1]})

    x = {"lat":coords4326[1], "long": coords4326[0]}
    conn.send(JSON.stringify(x));
};

document.getElementById("submitButton").onclick = function () {
    submitGuess();
    // console.log("Send ready message!")
};

console.log("ws_related loaded successfully")

function addSolution(message){
    solution = new ol.Feature({
        geometry: new ol.geom.Point(ol.proj.transform([message.lng, message.lat], 'EPSG:4326', 'EPSG:3857'))
        });

    solution_layer.getSource().addFeature(solution)
    flyTo(ol.proj.fromLonLat([message.lng, message.lat]), function (){});
}
