fetch("/data/default", {method: "GET", credentials: "same-origin"})
.catch((error) => {
  console.error('Error:', error);
}).then(data => { return data.json()}).then(defaultData => initIGV(defaultData))



function initIGV(defaultData) {
    var igvDiv = document.getElementById("igv-div-1");
    igv.createBrowser(igvDiv, defaultData)
    .then(function (browser) {
        igvBrowser = browser;
        console.log("Created IGV browser 1");
    })
}

function addBigWigsTrack(id) {
  var basePath = "/data/bigWigsTrack/"
  var fullPath = basePath + id
  fetch(fullPath, {method: "GET", credentials: "same-origin"})
  .catch((error) => {
  console.error('Error:', error);
}).then(data => { return data.json()}).then(tracks => addTrack(tracks))
}

function addBamTrack(id) {
  var basePath = "/data/bamTrack/"
  var fullPath = basePath + id
  fetch(fullPath, {method: "GET", credentials: "same-origin"})
  .catch((error) => {
  console.error('Error:', error);
}).then(data => { return data.json()}).then(tracks => addTrack(tracks))
}

function addTrack(tracks) {
  for (let track of tracks) {
    igvBrowser.loadTrack(track)
  }
}