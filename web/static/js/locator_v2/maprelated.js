var token = 'pk.eyJ1IjoidGltd2VuZGVsIiwiYSI6ImNraDBoeG9ubTFkd20zMXJydDA5YjR0OXEifQ.uuncQPQVAT2rDxfj81AILw';

const mb = new ol.layer.Tile({
    source: new ol.source.XYZ({
        // url: 'https://api.mapbox.com/styles/v1/mapbox/satellite-v9/tiles/256/{z}/{x}/{y}?access_token=' + token
        url: 'https://api.mapbox.com/v4/mapbox.satellite/{z}/{x}/{y}.png?access_token=' + token
    }),
});



const fill = new ol.style.Fill({
    color: 'rgba(255, 255, 0, 0.1)',
});

const stroke = new ol.style.Stroke({
    color: "#FF0000",
    width: 1.25,
});

const fill_submits = new ol.style.Fill({
    color: 'rgba(100, 100, 100, 0.1)',
});

const stroke_submits = new ol.style.Stroke({
    color: "#08ff00",
    width: 1.25,
});

const submit_Style = new ol.style.Style({
    image: new ol.style.Circle({
        fill: fill_submits,
        stroke: stroke_submits,
        radius: 1,
    }),
    text: new ol.style.Text({
        offsetY : 15,
        font: 'bold 17px "Open Sans", "Arial Unicode MS", "sans-serif"',
        // placement: 'line',
        fill: new ol.style.Fill({
            color: 'black',
        }),
        stroke : new ol.style.Stroke({
            color: "#FFFFFF",
            width: 1.25,
        }),
    }),
})

const styles = [
    new ol.style.Style({
        image: new ol.style.Circle({
            fill: fill,
            stroke: stroke,
            radius: 5,
        }),
        fill: fill,
        stroke: stroke,
    }),
];

point = new ol.Feature({
    geometry: new ol.geom.Point(ol.proj.transform([0,0], 'EPSG:4326', 'EPSG:3857'))
});

var layer = new ol.layer.Vector({
    source: new ol.source.Vector({
        features: [point]
    })
});

var solution_layer = new ol.layer.Vector({
    source: new ol.source.Vector({
        features: []
    }),
    style: styles[0]
});

var submit_Layer = new ol.layer.Vector({
    source: new ol.source.Vector({
        features: []
    }),
    style: function (feature) {
        submit_Style.getText().setText(feature.get('name'));
        return submit_Style;
    },
});

const select = new ol.interaction.Select({
    layers : [layer],
});

const translate = new ol.interaction.Translate({
    features: select.getFeatures(),
});

const map = new ol.Map({
    interactions: ol.interaction.defaults().extend([select, translate]),
    layers: [mb],
    target: 'map',
    // view: new ol.View({
    //     center: [0, 0],
    //     zoom: 1,
    //     }),
    view: new ol.View({
        center: ol.proj.transform(center, 'EPSG:4326', 'EPSG:3857'),
        zoom: zoom,
        maxZoom: maxZoom,
        minZoom: minZoom,
        extent: ol.proj.transformExtent(extent, 'EPSG:4326', 'EPSG:3857'),
    }),
});



map.addLayer(layer);
map.addLayer(solution_layer);
map.addLayer(submit_Layer);

/**
 * Elements that make up the popup.
 */
const popupDIV = document.getElementById('popup');
const content = document.getElementById('popup-content');
const closer = document.getElementById('popup-closer');
const addInfos = document.getElementById('addInfos');
/**
 * Create an overlay to anchor the popup to the map.
 */
const popup = new ol.Overlay({
    element: popupDIV,
    positioning: 'top-left',
    //  autoPan: {
    //    animation: {
    //      duration: 250,
    //    },
    //  },
});

closer.onclick = function () {
    popup.setPosition(undefined);
    closer.blur();
    return false;
};
map.addOverlay(popup);

map.on('click', function(evt){
    //console.log(ol.proj.transform(evt.coordinate, 'EPSG:3857', 'EPSG:4326'));
    point.setGeometry(new ol.geom.Point(evt.coordinate))
});

map.on("click", function(){
    popup.setPosition(undefined);
    closer.blur();
})

addInfos.onclick = function(){
    if (message.status=="reviewing"){
        popup.setPosition(ol.proj.transform([message.lng, message.lat], 'EPSG:4326', 'EPSG:3857'))
        return
    }
    popup.setPosition(map.getView().getCenter())
}

const notifierDIV = document.getElementById('notifier');
const notifierContent = document.getElementById('notifiercontent');
const notifierOL = new ol.Overlay({
    element: notifierDIV,
    positioning: 'center-center'
});

map.addOverlay(notifierOL);

function flyTo(location, done) {
    const duration = 2000;
    const zoom = map.getView().getZoom();
    let parts = 2;
    let called = false;
    function callback(complete) {
        --parts;
        if (called) {
            return;
        }
        if (parts === 0 || !complete) {
            called = true;
            done(complete);
        }
    }
    map.getView().animate(
        {
            center: location,
            duration: duration,
        },
        callback
    );
    map.getView().animate(
        {
            zoom: zoom - 1,
            duration: duration / 2,
        },
        {
            zoom: zoom,
            duration: duration / 2,
        },
        callback
    );
};

console.log("maprelated loaded successfully")