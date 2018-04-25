<!-- Initializing Senna -->
<!-- TODO Maybe move this to a JS file -->
'use strict';
var calling_path;
var app = new senna.App();
app.setBasePath('/ui');
app.addSurfaces([ 'pbconf_surface']);

app.addRoutes(new senna.Route(/\?home$/, function() {
	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			showView("#home-view");
			showNavHome();
			return "";
		}
	}
	return screen;
}));

app.addRoutes(new senna.Route(/\?nodes$/, function(event) {
	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			showView("#node-view");
			showNavNodes();
			return "";
		}
	}
	return screen;
}));
app.addRoutes(new senna.Route(/\?devices$/, function() {
	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			showView("#device-view");
			showNavDevices();
			return "";
		}
	}
	return screen;
}));
app.addRoutes(new senna.Route(/\?policies$/, function() {
	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			showView("#policies-view");
			showNavPolicies();
			return "";
		}
	}
	return screen;
}));
app.addRoutes(new senna.Route(/\?reports$/, function() {
	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			showView("#reports-view");
			showNavReports();
			return "";
		}
	}
	return screen;
}));
app.addRoutes(new senna.Route(/\?node-*/, function() {
	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			var matches = calling_path.match(/node-(.*)__(.*)/);
			showNodeDetail(matches[1], matches[2]);
			showNavNodes();
			return "";
		}
	}
	return screen;
}));

app.addRoutes(new senna.Route(/\?device-.*/, function() {

	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			var matches = calling_path.match(/device-(.*)__(.*)/);
			showDeviceDetail(matches[1],matches[2]);
			showNavDevices();
			return "";
		}
	}
	return screen;
}));

app.addRoutes(new senna.Route(/\?policy-.*/, function() {

	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			var decoded_url = decodeURIComponent(calling_path);
			var matches = decoded_url.match(/policy-(.*)/);
			launchPolicyPage(matches[1]);
			showNavPolicies();
			return "";
		}
	}
	return screen;
}));

app.addRoutes(new senna.Route(/\?report-.*/, function() {

	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			var decoded_url = decodeURIComponent(calling_path);
			var matches = decoded_url.match(/report-(.*)/);
			launchReportPage(matches[1])
			showNavReports();
			return "";
		}
	}
	return screen;
}));

app.addRoutes(new senna.Route(/\?deviceconfig-.*/, function() {

	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			var decoded_url = decodeURIComponent(calling_path);
			var matches = decoded_url.match(/deviceconfig-(.*)/);
			showDeviceConfig(matches[1]);
			showNavDevices();
			return "";
		}
	}
	return screen;
}));


app.addRoutes(new senna.Route(/\?devicemetadata-.*/, function() {

	var screen = new senna.Screen();
	screen.getSurfaceContent = function(surfaceId) {
		if (surfaceId === 'pbconf_surface') {
			var decoded_url = decodeURIComponent(calling_path);
			var matches = decoded_url.match(/devicemetadata-(.*)__(.*)/);
			showDeviceMetadata(matches[1],matches[2]);
			showNavDevices();
			return "";
		}
	}
	return screen;
}));

// Get the URL that's been called
app.on('startNavigate', function(event) {
	calling_path = event.path;
});


// Handle initial page load to show appropriate view
function handleNavigation() {
	var loc = window.location.toString();
	var params = loc.split("?");
	if (params.length == 2) {
		app.navigate('/ui?' + params[1]);
	}
}

// Functions below are workarounds to highlight the
// correct item on the navigation bar at the top

function showNavHome() {
	document.getElementById('home').className='active';
	document.getElementById('policies_link').className='';
	document.getElementById('nodes_link').className='';
	document.getElementById('reports_link').className='';
	document.getElementById('devices_link').className='';
	var container = document.getElementById('navbar');
	var refreshContent = container.innerHTML;
	container.innerHTML = refreshContent;
}

function showNavNodes() {
	document.getElementById('home').className='';
	document.getElementById('policies_link').className='';
	document.getElementById('devices_link').className='';
	document.getElementById('reports_link').className='';
	document.getElementById('nodes_link').className='active';
	var container = document.getElementById('navbar');
	var refreshContent = container.innerHTML;
	container.innerHTML = refreshContent;
}

function showNavDevices() {
	document.getElementById('home').className='';
	document.getElementById('policies_link').className='';
	document.getElementById('nodes_link').className='';
	document.getElementById('reports_link').className='';
	document.getElementById('devices_link').className='active';
	var container = document.getElementById('navbar');
	var refreshContent = container.innerHTML;
	container.innerHTML = refreshContent;
}

function showNavPolicies() {
	document.getElementById('policies_link').className='active';
	document.getElementById('home').className='';
	document.getElementById('nodes_link').className='';
	document.getElementById('reports_link').className='';
	document.getElementById('devices_link').className='';
	var container = document.getElementById('navbar');
	var refreshContent = container.innerHTML;
	container.innerHTML = refreshContent;
}

function showNavReports() {
	document.getElementById('policies_link').className='';
	document.getElementById('home').className='';
	document.getElementById('nodes_link').className='';
	document.getElementById('reports_link').className='active';
	document.getElementById('devices_link').className='';
	var container = document.getElementById('navbar');
	var refreshContent = container.innerHTML;
	container.innerHTML = refreshContent;
}
