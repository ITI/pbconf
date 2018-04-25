$(document).ready(function(){
    buildNodeTable(); //not necessary but nice to see something instantly
    buildDeviceTable();
    buildPoliciesTable();
    buildReportsTable();
    // Build tree with nodes and devices
    buildDashboard();
    // Build dashboard pane with reports
    buildDashReports();
    // Store times to prevent unnecessary/expensive dashboard refresh
    initializeDashTimer();

    if ($("#EnableLogPoll").text() == "true") {
        startAlarmLongPoll();
    }
    updateTimer = setInterval(onTimerTick, 6000);
});
/***************** global variables *****************************/
var updateTimer;
var modifiedTag  = {"Nodes":"", "Devices":"", "Policy":"", "Report":""};
var modifiedTagDash  = {"Nodes":"", "Devices":"", "Policy":"", "Report":""};
var deviceCfgFiles;
var singlePolicyFiles;

/*** Global variables for dashboard ***/
// array to store errors
var alarm_array = [];
// track nodes
var dashboard_nodes = [];
// track devices
var dashboard_devices = [];
// track state of treeview
var expanded_tree_nodes = [];

/***************** wrapper for ajax calls *****************************/
$(document).ajaxComplete(function(event, jqXHR, ajaxOptions){
    if (jqXHR.status == 401) {
        $.ajax({
            url: "/ui",
            method:"GET"
        }).done(function(newContent){
            clearInterval(updateTimer);
            document.open();
            document.write(newContent);
            document.close();
        });
    }
});
//Moves the highlighting of the item in the navigation bar Bootstrap specific
$(".nav a").on("click", function(){
   $(".nav").find(".active").removeClass("active");
   $(this).parent().addClass("active");
});

/*** Poll for alarms ***/
function startAlarmLongPoll() {
    $.ajax({
        url:"/alarm",
        method:"GET",
    //    timeout: 60000, //leaving this out causes browser to crash when pbconf is shut down.
    }).done(function(data, textStatus, jqXHR){
        startAlarmLongPoll();
        alarm_array.push(data);
        // Add error to dashboard
        $("#alarm-list form table").append("<tr><td><div class=\"checkbox\" id=\"prob_checkdiv\">" +
        "<label><input type=\"checkbox\" value=\"\">" + data + "</label></div></td></tr>");
        setAlarmVisibility();
    }).fail(function(jqXHR, textStatus, errorThrown){
        if (textStatus == "timeout" || jqXHR.status == 408){
            startAlarmLongPoll();
        }
    });
}

function onTimerTick(){
    //send HEAD requests to nodes, devices
    switch (currentView()) {
        case "#node-view":
            $.ajax({
                url:"/node",
                method:"HEAD",
            }).done(function(data, textStatus, jqXHR){
                nodesLastModified = jqXHR.getResponseHeader("Last-Modified");
                if (modifiedTag.Nodes == "" || modifiedTag.Nodes != nodesLastModified){
                    buildNodeTable();
                    modifiedTag.Nodes = nodesLastModified;
                }
            });
            $('#node-view td:first-child:gt(0)').each(function(index, element){
                node_ts_id = $(this).text(); //cache it to use later in .done, passed via context
                $.ajax({
                    url:"/node/"+ node_ts_id + "/Checkin Timestamp",
                    method:"GET",
                    context:{nodeId:node_ts_id}
                }).done(function(data){
                    var tstamp = $.parseJSON(data);
                    $('#TimeStamp' + this.nodeId).html(tstamp.Value);
                 });
              });
            break;
        case "#device-view":
            $.ajax({
                url:"/device",
                method:"HEAD",
            }).done(function(data, textStatus, jqXHR){
                devicesLastModified = jqXHR.getResponseHeader("Last-Modified");
                if (modifiedTag.Devices == "" || modifiedTag.Devices != devicesLastModified){
                    buildDeviceTable();
                    modifiedTag.Devices = devicesLastModified;
                }
            });
            break;
        case "#policies-view":
        $.ajax({
            url:"/policy",
            method:"HEAD"
        }).done(function(data, textStatus, jqXHR){
            policiesLastModified = jqXHR.getResponseHeader("X-Pbconf-Policy-LastCommitId");
            if (modifiedTag.Policy == "" || modifiedTag.Policy != policiesLastModified){
                buildPoliciesTable();
                modifiedTag.Policy = policiesLastModified;
            }
        });
        break;
        //If user is on the report page (not table), update the contents if the HEAD request indicates change.
        case "#reportFile-view":
        $.ajax({
            url:"/reports",
            method:"HEAD"
        }).done(function(data, textStatus, jqXHR){
            reportsLastModified = jqXHR.getResponseHeader("X-Pbconf-Report-LastCommitId");
            if (modifiedTag.Report != reportsLastModified){
                ajaxGetReportContent();
                modifiedTag.Report = reportsLastModified;
            }
        });
        
        // Dashboard View
        case "#home-view":
        	// First check whether the nodes have changed
            $.ajax({
                url:"/node",
                method:"HEAD",
            }).done(function(data, textStatus, jqXHR){
                nodesLastModified = jqXHR.getResponseHeader("Last-Modified");
                // If so, build tree
                if (modifiedTagDash.Nodes == "" || modifiedTagDash.Nodes != nodesLastModified){
                	buildDashboard();
                	// update node updated times
                    modifiedTagDash.Nodes = nodesLastModified;
                }
            });

            // Check to see whether devices have changed
            $.ajax({
                url:"/device",
                method:"HEAD",
            }).done(function(data, textStatus, jqXHR){
            	devicesLastModified = jqXHR.getResponseHeader("Last-Modified");
                // If so, build tree
                if (modifiedTagDash.Devices == "" || modifiedTagDash.Devices != devicesLastModified){
                	buildDashboard();
                 // Also update device updated times
                    modifiedTagDash.Devices = devicesLastModified;
                }
            });

            // Also check to see whether reports have changed
            $.ajax({
                url:"/reports",
                method:"HEAD"
            }).done(function(data, textStatus, jqXHR){
                reportsLastModified = jqXHR.getResponseHeader("X-Pbconf-Report-LastCommitId");
                // If so, add to reports pane
                if (modifiedTagDash.Report != reportsLastModified){
                	buildDashReports();
                    modifiedTagDash.Report = reportsLastModified;
                }
            });
        break;

    }
}

function setRoleBasedOptions(){
    if ($("#UserRole").text() != "admin"){
        //devicetable icons
        $('button[id=editDevice]').addClass('hidden');
        $('button[id=deleteDevice]').addClass('hidden');
        //node config items table icons
        $('button[id=editCfgItem]').addClass('hidden');
        $('button[id=deleteCfgItem]').addClass('hidden');
        //device config items table icons
        $('button[id=editDeviceCfgItem]').addClass('hidden');
        $('button[id=deleteDeviceCfgItem]').addClass('hidden');
        //policies table icons
        $('button[id=editPolicyFile]').addClass('hidden');
        $('button[id=deletePolicyFile]').addClass('hidden');
        //device configurations table icons
        $('button[id=editDeviceCfgFile]').addClass('hidden');
        $('button[id=deleteDeviceCfgFile]').addClass('hidden');

        //hide the divs that show text box, Add New button
        //node, node config item
        $("#add-new-node").addClass('hidden');
        $("#add-new-cfgItem").addClass('hidden');
        //device, device config item
        $("#add-new-device").addClass('hidden');
        $("#add-new-devCfgItem").addClass('hidden');
        //policy, policy file update
        $("#add-new-policy").addClass('hidden');
        $("#update-policyFile").addClass('hidden');
        //reports, hide the Add New report button, and all the buttons (save, cancel, run now) in the update-reportFile group
        $("#add-new-report").addClass('hidden');
        $("#update-reportFile").addClass('hidden');
    }
}

/********************* Show/Hide views functions *************************/
// The functions for when tabs are clicked
$("#home").click(function(){
    showView("#home-view");
});
$("#nodes").click(function(){
    showView("#node-view");
    });
$("#devices").click(function(){
    showView("#device-view");
    });
$("#policies").click(function(){
    showView("#policies-view");
    });
$("#reports").click(function(){
    showView("#reports-view");
});
//single node config (items), dev config (items), dev configurations, policy detail views
function showConfigView(){
    showView("#config-view");
}
function showDeviceConfigView(){
    showView("#device-config-view");
}
//single file detail views
function showDeviceConfigFileView(){
    showView("#device-configFile-view");
}
function showPolicyFileView(){
    showView("#policyFile-view");
}
function showReportFileView(){
    showView("#reportFile-view");
}
function showDeviceMetaFileView(){
    showView("#device-metaFile-view");
}

//remove all alerts that may be showing
function removeAlert(){
    $(".alert").remove();
}

/*** Determines which view should be visible ***/
function showView(viewId){
    listViews = ["#home-view", "#node-view", "#device-view", "#policies-view", "#reports-view", "#config-view", "#device-config-view",
                 "#device-configFile-view", "#policyFile-view", "#reportFile-view", "#device-metaFile-view" ];
    //hide all the views
    for (i=0; i< listViews.length; i++){
        $(listViews[i]).addClass("hidden");
    }
    //show the selected one
    if (viewId != "")
        $(viewId).removeClass("hidden");
    removeAlert();
}

/*** Returns view that is currently visible ***/
function currentView(){
    listViews = ["#home-view", "#node-view", "#device-view", "#policies-view", "#reports-view", "#config-view", "#device-config-view",
                 "#device-configFile-view", "#policyFile-view", "#reportFile-view", "#device-metaFile-view" ];
    for (i=0; i< listViews.length; i++){
        if (! $(listViews[i]).hasClass("hidden")){
            return listViews[i];
        }
    }
}

/********* utility to show and hide the buttons in the table row ***********/
function displayButtons(element, showList, hideList){
    for (i=0; i < showList.length; i++){
        $(element).parent().find(showList[i]).removeClass('hidden');
    }
    for (j=0; j < hideList.length; j++){
        $(element).parent().find(hideList[j]).addClass('hidden');
    }
}

/********************* logout functions *************************/
function onLogout(){
    clearInterval(updateTimer);
    $.post("/ui/logout");
}

/********************* Node view functions *************************/
function onAddNode(){
    node = $("#add-new-node input[name=Name]");

    data = '{"'+ $(node).attr("name") + '":"' + $(node).val() + '"}';
    $.post("/node", data, function(){
        buildNodeTable();
    });
    $("#add-new-node input").val("");
}

function buildNodeTable(){
    //clear off old node items from the node table
    $("#node-view table tbody tr").remove();

    $.get("/node", function(data){
        var nodes = $.parseJSON(data);
        for (i=0;i< nodes.length; i++){
            row = $('<tr></tr>');
            row.append($('<td>' + nodes[i].Id + '</td>'));

            cell = $('<td><a href=' + '"?node-' + nodes[i].Name + '__' + nodes[i].Id  + '">'+nodes[i].Name + "</td>");
            cell.attr('id', 'node-detail');
            row.append(cell);

            row.append($('<td id=TimeStamp'+nodes[i].Id + '>' + '</td>'));
            $("#node-view table tbody").append(row);
        }
        setRoleBasedOptions();
        if (nodes.length > 0){
            $("#add-new-node").addClass('hidden');
        }
    });
}

function showNodeDetail(nodeName, nodeId) {
	$("#config-view").find("h3").replaceWith('<h3>Configuration items for ' + decodeURIComponent(nodeName) + '<small></h3>');
    buildConfigTable(nodeId);
    showConfigView();
}

/********************* Device view functions *************************/

// Populate table of devices
function buildDeviceTable(){
    //clear off the device table
    $("#device-view table tbody tr").remove();

    $.ajax({
        url: "/device", //Example: /v2/device to send version in the url
        method: "GET",
        /* Example: To send version in the accept header
        headers:{
            Accept: "application/json;pbconfversion=2"
        }
        */
    }).done(function(data){
        var devices = $.parseJSON(data);
        for(i=0; devices != null && i<devices.length;i++){
            row = $('<tr></tr>');
            row.append($('<td>' + devices[i].Id + '</td>'));

            cell = $('<td><a href=' + '"?device-' + devices[i].Name + '__' + devices[i].Id  + '">'+devices[i].Name + "</td>");
            cell.attr('id', 'device-detail');
            row.append(cell);

            row.append($('<td>' + devices[i].ParentNode + '</td>'));
            row.append($('<td class="text-right">' +
                                '<button type="button" id="editDevice" class="btn btn-default" aria-label="Edit"><span class="glyphicon glyphicon-pencil" aria-hidden="true"</span></button>'
                                + '<button type="button" id="doneDevice" class="btn  btn-default hidden" aria-label="Done"><span class="glyphicon glyphicon-ok" aria-hidden="true"</span></button>'
                                + '<button type="button" id="cancelDevice" class="btn  btn-default hidden" aria-label="Cancel"><span class="glyphicon glyphicon-remove" aria-hidden="true"</span></button>'
                                + '<button type="button" id="deleteDevice" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                                +'</td>'));
            $("#device-view table tbody").append(row);
        }
        setRoleBasedOptions();
    });
}

// Delete device
$('#device-view').on('click', '#deleteDevice', function(event){
    deviceRow = $(this).parent().parent();
    deviceId = $(deviceRow).find("td:first").text();
    $.ajax({
        url: "/device/" + deviceId,
        method: "DELETE"
    }).done(function(){
        buildDeviceTable();
    });
});

// Enable editing of device
$('#device-view').on('click', '#editDevice', function(event){
    displayButtons(this, ['#doneDevice', '#cancelDevice'], ['#editDevice']);
    devElement = $(this).parent().parent().find('td:nth-child(2)');
    devElement.html('<input type = "text" value="' + devElement.text() +  '">');
    devElement.removeAttr("id");  //remove ability of clicking on name leading to showing config items for device
});

// Finish editing device
$('#device-view').on('click', '#doneDevice', function(event){
    displayButtons(this, ['#editDevice'], ['#doneDevice', '#cancelDevice']);
    devId = $(this).parent().parent().find('td:nth-child(1)').text();
    devElement = $(this).parent().parent().find('td:nth-child(2)');
    devParentId = $(this).parent().parent().find('td:nth-child(3)').text();
    devName = devElement.find('input').attr("value");

    data = '{"Name": "'+ $(devElement).find('input').val() + '"' + ', "ParentNode": ' + devParentId + ', '+ '"ConfigItems": []}';

    $.ajax({
        url:"/device/" + devId,
        method:"PATCH",
        context:devElement,
        data:data
    }).done(function(){
        $(this).html('<a href="?device-' + devName + '__' + devId  + '">' + $(this).find('input').val() + '</a>');
        $(this).attr('id', 'device-detail');
    })
    .fail(function(){
        $(this).html('<a href="?device-' + devName + '__' + devId  + '">' + $(this).find('input').attr("value") + '</a>');
        $(this).attr('id', 'device-detail');
    });
});

// Cancel editing device without saving changes
$('#device-view').on('click', '#cancelDevice', function(event){
    displayButtons(this, ['#editDevice'], ['#doneDevice', '#cancelDevice']);
    dElem = $(this).parent().parent().find('td:nth-child(2) input');
    devName = $(dElem).attr("value");
    devId = $(this).parent().parent().find('td:nth-child(1)').text();
    dElem.parent().html('<a href="?device-' + devName + '__' + devId  + '">' + $(dElem).attr("value") + '</a>');
    $(this).parent().parent().find('td:nth-child(2)').attr('id', 'device-detail'); 

});

function onAddDevice(){
    device = $("#add-new-device input[name=Name]");
    data = '{' + '"'+ $(device).attr("name") + '":"' + $(device).val() + '", '+ '"ConfigItems": []}';
    $.post("/device", data, function(){
        buildDeviceTable();
    });
    // Reset new device name to blank
    $("#add-new-device input").val("");
    // Disable "Add New" button
    $("#new-device-btn").prop("disabled", true);
}

function showDeviceDetail(deviceName, deviceId) {
	$("#device-config-view").find("h3").replaceWith('<h3>Configuration items for ' + decodeURIComponent(deviceName) + '<small></h3>');
    buildDeviceConfigTable(deviceId);
    showDeviceConfigView();
}

/********************* Policies view functions *************************/

// Populate table with policies
function buildPoliciesTable(){
    //clear off old policy items from the policy table
    $("#policies-view table tbody tr").remove();

    $.get("/policy", function(data){
        var policyList = $.parseJSON(data);
        for (i=0;i< policyList.length; i++){
            addRowPoliciesTable(policyList[i]);
        }
        setRoleBasedOptions();
    });
}

// Add new policy to table
function addRowPoliciesTable(policyName){
    row = $('<tr></tr>');
    cell = $('<td><a href="?policy-' + policyName + '">'+ policyName + "</td>");
    cell.attr('id', 'policy-detail');
    row.append(cell);
    row.append($('<td class="text-right">'
                        + '<button type="button" id="deletePolicyFile" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                        +'</td>'));
    $("#policies-view table tbody").append(row);
}

// Delete policy file
$('#policies-view').on('click', '#deletePolicyFile', function(event){
    policyRow = $(this).parent().parent();
    policyName = $(policyRow).find("td:first").text();
    $.ajax({
        url: "/policy/" + policyName,
        method: "DELETE"
    }).done(function(){
        buildPoliciesTable();
    });
});

// After adding policy, open view to edit new policy
function onAddPolicy(){
    newPolicyName = $("#add-new-policy input[name=Name]").val();
    setRoleBasedOptions();
    // Reset value to blank
    $("#add-new-policy input").val("");
    // Disable add button
    $("#new-policy-btn").prop("disabled", true);
    app.navigate('/ui?policy-' + newPolicyName);
}
/********************* Reports view functions *************************/

// Populate table of reports
function buildReportsTable(){
    //clear off old policy items from the policy table
    $("#reports-view table tbody tr").remove();

    $.get("/reports", function(data){
        var reportList = $.parseJSON(data);
        for (i=0; i<reportList.length; i++){
            addRowReportsTable(reportList[i]);
        }
        setRoleBasedOptions();
    });
}

// Add new report to table
function addRowReportsTable(reportName){
    row = $('<tr></tr>');
    cell = $('<td><a href="?report-'+ reportName +'">'+ reportName + "</td>");
    cell.attr('id', 'report-detail');
    row.append(cell);
    row.append($('<td class="text-right">'
                        + '<button type="button" id="deleteReportFile" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                        +'</td>'));
    $("#reports-view table tbody").append(row);
}

// Delete report file
$('#reports-view').on('click', '#deleteReportFile', function(event){
    reportRow = $(this).parent().parent();
    reportName = $(reportRow).find("td:first").text();
    $.ajax({
        url: "/reports/" + reportName,
        method: "DELETE"
    }).done(function(){
        buildReportsTable();
    });
});

function launchReportPage(reportName){
	$("#reportFile-view").find("h4").replaceWith('<h4>Report Query file ' + decodeURIComponent(reportName) + '<small></h4>');
    buildReportFileView(reportName);
    showReportFileView();
}

// After adding report, navigate to screen to edit report
function onAddReport(){
    newReportName = $("#add-new-report input[name=Name]").val();
    setRoleBasedOptions();
    // Reset value to blank
    $("#add-new-report input").val("");
    // Disable add button
    $("#new-report-btn").prop("disabled", true);
    app.navigate('/ui?report-' + newReportName);
}

// View to build new report
function buildReportFileView(reportName){
    removeAlert();
    //clear off old report File shown in div?
    $("#reportFile-query").remove(); //remove any query content
    $("#period-reportFile input").val(""); //blank period
    $('#period-reportFile input[name="periodic-checkbox"]').prop('checked', false); //uncheck checkbox
    $('#reportFile-content').replaceWith('<div id="reportFile-content" class="col-xs-12"></div>');
    // Initially disable export button
    $("#rpt-button-export").prop("disabled", true);
    $("#reportFile-view").attr("reportName", reportName);
    content = "";
    $("#reportFile-view").attr('reportTimeStamp', new Date().toUTCString()); //current time
    $.get("/reports/"+ reportName, function(data){
        if (data != "") {
            queryFile = $.parseJSON(data);
            content = queryFile.Query;
            $("#reportFile-view").attr('reportTimeStamp', queryFile.TimeStamp); //replace with time in report
            if (queryFile.Period != -1){
                $('#period-reportFile input[name="periodic-checkbox"]').prop('checked', true);
                $("#period-reportFile input").val(queryFile.Period);
            }
        }
    })
    .always(function(){
    	$('#reportFile-view h4').after($('<div id="reportFile-query"><textarea id="queryText" style="width: 100%; height: 150px;">' + decodeURIComponent(content) +  '</textarea></div>'));
    	// Enable run button if there is a query
    	if (content.length > 0) {
        	//$("#runReport").prop("disabled",false);
    		enableRunReportButtons(true);
    	}
    	else {
    		// If there's no query, disable all buttons inititally
    		disableAllButtons();
    	}
    	setRoleBasedOptions();
        // Initially disable export?
    });
    ajaxGetReportContent();
}

// When user clicks "Save Changes" button, save to server
function onSaveReportFile(){
    timestamp = $("#reportFile-view").attr('reportTimeStamp'); // use cached time in report. This will not cause a run
    msg = $("#msg-reportFile input[name=report-commit-msg]").val();
    reportName = $("#reportFile-view").attr("reportName");
    var exists = reportExists(reportName)
    if (exists == true) {
        ajaxPutReport(timestamp, msg, true);
    }
    else {
        ajaxPostReport(timestamp, msg);
    }
}

// Does a report with this name already exist?
function reportExists(reportName) {
    var exists = false
    $("#reports-view table tbody tr").each(function(){
        row = $(this);
        if (row.find("td:first").text() == reportName) {
            exists = true;
        }
    });
    return exists
}

function ajaxPostReport(timestamp, commitMessage) {
    removeAlert();
    data = getReportData(timestamp, commitMessage)

    $.ajax({
        url:"/reports",
        method:"POST",
        contentType: "application/json; charset=utf-8",
        data:data
    })
    .done(function(){
        $("#msg-reportFile input").val("");
        $( '<div class="alert alert-success" role="alert">Successfully committed the report query file.</div>').insertAfter( "#update-reportFile" );
        ajaxGetReportContent();
        enableRunReportButtons(true);
        addRowReportsTable(reportName);
    })
    .fail(function(){
        $("#msg-reportFile input").val("");
        $('<div class="alert alert-danger" role="alert">Failed to create the report query file.</div>').insertAfter( "#update-reportFile" );
    });
}

function getReportData(timestamp, commitMessage) {
    reportName = $("#reportFile-view").attr("reportName");
    content = $("#reportFile-query textarea").val();
    periodValue = "-1";
    if ($('#period-reportFile input[name="periodic-checkbox"]').is(':checked')){
        periodValue = $('#period-reportFile input[name="period-text"]').val();
    }
    data = '{' +
        '"Name": ' +
        '"' + reportName + '"' +
        ',' +
        '"Query" : ' +
        '"' + content + '"' +
        ',' +
        '"Period" : ' +
        '"' + periodValue + '"' +
        ',' +
        '"TimeStamp" : ' +
        '"' + timestamp + '"' +
        ',' +
        '"Author": ' +
        '"' + $('#UserName').text() + '"' +
        ',' +
        '"CommitMessage": ' +
        '"' + commitMessage + '"' +
        '}';
        return data;
}

function ajaxPutReport(timestamp, commitMessage, displayAlert) {
    removeAlert();
    data = getReportData(timestamp, commitMessage)

    $.ajax({
        url:"/reports/" + reportName,
        method:"PUT",
        contentType: "application/json; charset=utf-8",
        data:data
    })
    .done(function(){
        $("#msg-reportFile input").val("");
        if (displayAlert){
            $( '<div class="alert alert-success" role="alert">Successfully committed changes to the report query file.</div>').insertAfter( "#update-reportFile" );
        }
        ajaxGetReportContent();
        // Only disable save/cancel if save is successful?
        enableRunReportButtons(true);
    })
    .fail(function(){
        $("#msg-reportFile input").val("");
        if (displayAlert){
            $('<div class="alert alert-danger" role="alert">Failed to update the report query file.</div>').insertAfter( "#update-reportFile" );
        }
    });
}

function ajaxGetReportContent(){
    reportName = $("#reportFile-view").attr("reportName");
    $.ajax({
        url:"/reports/report/" + reportName,
        method:"GET",
    })
    .done(function(data){
    	$('#reportFile-content').replaceWith('<div id="reportFile-content" class="col-xs-12">' + reportResultsTable(data) +  '</div>');
    	Sortable.init();
    	// Enable export button if there's data to export
    	$("#rpt-button-export").prop("disabled", data == "");
    });
}

function onCancelReportFile(){
    removeAlert();
    reportName = $("#reportFile-view ").attr('reportName');
    $("#msg-reportFile input").val("");
    buildReportFileView(reportName);
    showReportFileView();
}

function onRunReportFile(){
    timestamp = new Date().toUTCString(); //replace cached time with time now.
    commitMessage = "Report was manually run at " + timestamp;
    ajaxPutReport(timestamp, commitMessage, false);
}

$('#reportFile-view').on("change keyup paste", "#reportFile-query", function(){
	// Check value of query
	enableRunReportButtons(false);
});

$('#reportFile-view').on("change keyup paste", "#reportFile-query-period", function(){
	// Check value of query
	enableRunReportButtons(false);
});

// Handle enabling/disabling save/cancel buttons
function enableRunReportButtons(flag) {
	// Disabled by default
	$("#rpt-button-save").prop("disabled", flag);
	$("#rpt-button-cancel").prop("disabled", flag);
	$("#runReport").prop("disabled", !flag);
}

function disableAllButtons() {
	$("#rpt-button-save").prop("disabled", true);
	$("#rpt-button-cancel").prop("disabled", true);
	$("#runReport").prop("disabled", true);
}


/********************* Policy File view functions *************************/
function launchPolicyPage(policyName){
    buildPolicyFileView(policyName);
    showPolicyFileView();
}

function buildPolicyFileView(policyName){
    removeAlert();
    //clear off old policy File shown in div?
    $("#policyFile-content").remove();
    $("#policyFile-view").find("h4").replaceWith('<h4>Policy file ' + decodeURIComponent(policyName) + '<small></h4>');
    $('#policyFile-content').replaceWith('<div id="policyFile-content" class="col-md-6"></div>');
    $("#policyFile-view").attr("policyName", policyName);
    content = ""
    $.get("/policy/"+ policyName, function(data){
        var pol = $.parseJSON(data)

        $("#policyFile-view").attr("polName", policyName);
        content = pol.Files["Rules"]
    }).always(function(){
        $('#policyFile-view h4').after($('<div id="policyFile-content"><textarea style="width: 100%; height: 150px;">' + decodeURIComponent(content) +  '</textarea></div>'));
    });
}

function policyExists(policyName) {
    var exists = false
    $("#policies-view table tbody tr").each(function(){
        row = $(this);
        if (row.find("td:first").text() == policyName) {
            exists = true;
        }
    });
    return exists
}

function onSavePolicyFile(){
    removeAlert();
    policyName = $("#policyFile-view").attr('policyName');
    content = $("#policyFile-content textarea").val();
    msg = $("#msg-policyFile input[name=pol-commit-msg]").val();
    data = '{' +
        '"Content": ' + '{' +
        '"Files": {"Rules": "' + encodeURIComponent(content) + '"' +
        '}' +
        ',' +
        '"Object": "' + policyName + '"' +
        '}' + ', ' + //end of CMContent
        '"Author": ' + '{"Name": "' + $('#UserName').text() + '"}' + ', ' +
        '"Log": ' + '{' +
        '"Message": "' + msg + '"' + '}' +
        '}';

    var exists = policyExists(policyName)

    $.ajax({
        url:"/policy/" + policyName,
        method:"PUT",
        contentType: "application/json; charset=utf-8",
        data:data
    })
    .done(function(){
        $("#msg-policyFile input").val("");
        $( '<div class="alert alert-success" role="alert">Successfully committed changes to the policy file.</div>').insertAfter( "#update-policyFile" );
        if (exists == false) {
            addRowPoliciesTable(policyName);
        }
    })
    .fail(function(){
        $("#msg-policyFile input").val("");
        $('<div class="alert alert-danger" role="alert">Failed to commit changes to the policy file.</div>').insertAfter( "#update-policyFile" );
    });
}

function onCancelPolicyFile(){
    removeAlert();
    policyName = $("#policyFile-view").attr('policyName');
    $("#msg-policyFile input").val("");
    buildPolicyFileView(policyName);
    showPolicyFileView();
}
/********************* Config view functions *************************/

// Populate table for node config
function buildConfigTable(nodeId){
    //clear off old config items from the config table
    $("#config-view table tbody tr").remove();
    $.get("/node/"+nodeId, function(data){
        var node = $.parseJSON(data);
        $("#config-view table").attr("id", nodeId);
        $("#config-view table").attr("nodename", node.Name);
        configItems = node.ConfigItems;
        if (configItems == null){
            return;
        }
        for(i=0;i<configItems.length;i++){
            row = $('<tr></tr>');
            row.append($('<td>' + configItems[i].Key + '</td>'));
            //limit the config item value to 25 chars
            var cfgValue = configItems[i].Value;
            row.append($('<td style="word-wrap:break-word">' + cfgValue + '</td>'));
            row.append($('<td class="text-right">'
                                + '<button type="button" id="editCfgItem" class="btn btn-default" aria-label="Edit"><span class="glyphicon glyphicon-pencil" aria-hidden="true"</span></button>'
                                + '<button type="button" id="doneCfgItem" class="btn  btn-default hidden" aria-label="Done"><span class="glyphicon glyphicon-ok" aria-hidden="true"</span></button>'
                                + '<button type="button" id="cancelCfgItem" class="btn  btn-default hidden" aria-label="Cancel"><span class="glyphicon glyphicon-remove" aria-hidden="true"</span></button>'
                                + '<button type="button" id="deleteCfgItem" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                                +'</td>'));

            $("#config-view table tbody").append(row);
        }
        setRoleBasedOptions();
    });
}

// Add new node config item
function onAddCfgItem(){
    cfgKey = $("#add-new-cfgItem input[name=Key]");
    cfgValue = $("#add-new-cfgItem input[name=Value]");
    nodeId =  $("#config-view table").attr("id");
    nodeName =  $("#config-view table").attr("nodename");
    data = '{"Id":' + nodeId + ', "Name": "' + nodeName + '",' + '"ConfigItems" : ' +
        '[{"'+ $(cfgKey).attr("name") + '":"' + $(cfgKey).val() + '", "' + $(cfgValue).attr("name") + '" : "' + $(cfgValue).val() + '"}]' +
        '}';

    $.ajax({
        url:"/node/" + nodeId,
        method:"PATCH",
        data:data
    }).done(function(){
        buildConfigTable(nodeId);
    });
    // Reset value to blank
    $("#add-new-cfgItem input").val("");
    // Disable add button
    $("#new-config-btn").prop("disabled", true);
}

// Delete node config item
$('#config-view').on('click', '#deleteCfgItem', function(event){
    nodeId = $("#config-view table").attr("id");
    cfgRow = $(this).parent().parent();
    cfgKey = $(cfgRow).find("td:first").text();
    $.ajax({
        url: "/node/" + nodeId + "/" + cfgKey,
        method: "DELETE"
    }).done(function(){
        buildConfigTable(nodeId);
    });
});

// Enable editing for node config item
$('#config-view').on('click', '#editCfgItem', function(event){
    displayButtons(this, ['#doneCfgItem', '#cancelCfgItem'], ['#editCfgItem']);

    var valueElement = $(this).parent().parent().find('td:nth-child(2)');
    valueElement.html('<input type = "text" value="' + valueElement.text() +  '">');
});

// Finish editing node config item and save
$('#config-view').on('click', '#doneCfgItem', function(event){
    displayButtons(this, ['#editCfgItem'], ['#doneCfgItem', '#cancelCfgItem']);
    nodeId = $("#config-view table").attr("id");
    nodeName =  $("#config-view table").attr("nodename");
    cfgRow = $(this).parent().parent();
    cfgKey = $(cfgRow).find("td:first").text();
    cfgValueElem = $(cfgRow).find('td:nth-child(2) input');
    data = '{"Id":' + nodeId + ', "Name": "' + nodeName + '",' + '"ConfigItems" : ' +
        '[{"Key":"' + cfgKey + '", "Value" : "' + cfgValueElem.val() + '"}]' +
        '}';

    $.ajax({
        url:"/node/" + nodeId,
        method:"PATCH",
        data:data
    }).done(function(){
        cfgValueElem.parent().html(cfgValueElem.val());
    })
    .fail(function(){
        cfgValueElem.parent().html($(cfgValueElem).attr("value"));
    });
});

// Cancel editing node config item without saving changes
$('#config-view').on('click', '#cancelCfgItem', function(event){
    displayButtons(this, ['#editCfgItem'], ['#doneCfgItem', '#cancelCfgItem']);
    cfgValueElem = $(this).parent().parent().find('td:nth-child(2) input');
    cfgValueElem.parent().html($(cfgValueElem).attr("value"));
});

/********************* Device Config Items view functions *************************/
function buildDeviceConfigTable(deviceId){
    //clear off old config items from the config table
    $("#device-config-view table tbody tr").remove();
    $.get("/device/"+deviceId, function(data){
        device = $.parseJSON(data);
        var configItems = device.ConfigItems;
        $("#device-config-view table").attr("id", deviceId);
        $("#device-config-view table").attr("devName", device.Name);
        $("#device-config-view table").attr("parentNodeId", device.ParentNode);

        for(i=0; configItems != null && i<configItems.length;i++){
            row = $('<tr></tr>');
            row.append($('<td>' + configItems[i].Key + '</td>'));
            row.append($('<td style="word-wrap:break-word">' + configItems[i].Value + '</td>'));
            row.append($('<td class="text-right">'
                                + '<button type="button" id="editDeviceCfgItem" class="btn btn-default" aria-label="Edit"><span class="glyphicon glyphicon-pencil" aria-hidden="true"</span></button>'
                                + '<button type="button" id="doneDeviceCfgItem" class="btn  btn-default hidden" aria-label="Done"><span class="glyphicon glyphicon-ok" aria-hidden="true"</span></button>'
                                + '<button type="button" id="cancelDeviceCfgItem" class="btn  btn-default hidden" aria-label="Cancel"><span class="glyphicon glyphicon-remove" aria-hidden="true"</span></button>'
                                + '<button type="button" id="deleteDeviceCfgItem" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                                +'</td>'));

            $("#device-config-view table tbody").append(row);
        }
        //add row for config file link
        row = $('<tr></tr>');
        cell = $('<td><a href="?deviceconfig-' + deviceId +'">' + "Config File" + "</td>");
        cell.attr('id', 'device-cfgFiles-detail');
        row.append(cell);
        row.append($('<td style="word-wrap:break-word"></td>')); //col 2
        row.append($('<td style="word-wrap:break-word"></td>')); //col 3

        $("#device-config-view table tbody").append(row);
        //add row for Device meta data link
        if ($("#UserRole").text() == "admin"){
            row = $('<tr></tr>');
            cell = $('<td><a href="?devicemetadata-' + device.Name + '__' +  deviceId +'">' + "Metadata" + "</td>");
            cell.attr('id', 'device-meta-detail');
            row.append(cell);
            row.append($('<td style="word-wrap:break-word"></td>'));
            row.append($('<td style="word-wrap:break-word"></td>'));
            $("#device-config-view table tbody").append(row);
        }
        setRoleBasedOptions();
    });
}

function showDeviceConfig(deviceId) {
    deviceName =  $("#device-config-view table").attr("devName");
    $("#device-configFile-view").find("h4").replaceWith('<h4>Configuration file for ' + decodeURIComponent(deviceName) + '<small></h4>');
    buildConfigFileView(deviceId);
    showDeviceConfigFileView();
}

function showDeviceMetadata(deviceName, deviceId) {
	$("#device-metaFile-view").find("h4").replaceWith('<h4>Meta data for ' + decodeURIComponent(deviceName) + '<small></h4>');
    buildDeviceMetaTableView(deviceId);
    showDeviceMetaFileView();
}

function buildDeviceMetaTableView(deviceId){
    //clear off old config items from the config table
    $("#device-metaFile-view table tbody tr").remove();
    $.get("/device/"+deviceId + "/meta", function(data){
        var metaItems = $.parseJSON(data);
        $("#device-metaFile-view table").attr("id", deviceId);

        for (key in metaItems) {
            row = $('<tr></tr>');
            row.append($('<td>' + key + '</td>'));
            row.append($('<td style="word-wrap:break-word">' + metaItems[key] + '</td>'));
            row.append($('<td class="text-right">'
                                + '<button type="button" id="editDeviceMetaItem" class="btn btn-default" aria-label="Edit"><span class="glyphicon glyphicon-pencil" aria-hidden="true"</span></button>'
                                + '<button type="button" id="doneDeviceMetaItem" class="btn  btn-default hidden" aria-label="Done"><span class="glyphicon glyphicon-ok" aria-hidden="true"</span></button>'
                                + '<button type="button" id="cancelDeviceMetaItem" class="btn  btn-default hidden" aria-label="Cancel"><span class="glyphicon glyphicon-remove" aria-hidden="true"</span></button>'
                                + '<button type="button" id="deleteDeviceMetaItem" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                                +'</td>'));

            $("#device-metaFile-view table tbody").append(row);
        }
    });
}
function onAddDevMetaItem(){
    cfgKey = $("#add-new-devMeta input[name=Key]");
    cfgValue = $("#add-new-devMeta input[name=Value]");
    deviceId =  $("#device-metaFile-view table").attr("id");

    data = '{"'+ $(cfgKey).attr("name") + '":"' + $(cfgKey).val() + '", "' + $(cfgValue).attr("name") + '" : "' + $(cfgValue).val() + '"}';
    //data = '{"'+ $(cfgKey).val() +  '" : "' + $(cfgValue).val() + '"}'

    $.ajax({
        url:"/device/" + deviceId + "/meta",
        method:"PATCH",
        data:data
    }).done(function(){
        buildDeviceMetaTableView(deviceId);
    });
    // Reset value to blank
    $("#add-new-devMeta input").val("");
    // Disable add button
    $("#devmeta-att-btn").prop("disabled", true);
}

function onAddDevCfgItem(){
    cfgKey = $("#add-new-devCfgItem input[name=Key]");
    cfgValue = $("#add-new-devCfgItem input[name=Value]");
    deviceId =  $("#device-config-view table").attr("id");
    deviceName =  $("#device-config-view table").attr("devName");
    deviceParentId =  $("#device-config-view table").attr("parentNodeId");

    data = '{"Id":' +deviceId + ', ' +
        '"Name": "' + deviceName + '", ' +
        '"ParentNode": ' + deviceParentId + ', ' +
        '"ConfigItems" : ' +
        '[{"'+ $(cfgKey).attr("name") + '":"' + $(cfgKey).val() + '", "' + $(cfgValue).attr("name") + '" : "' + $(cfgValue).val() + '"}]' +
        '}';
    $.ajax({
        url:"/device/" + deviceId,
        method:"PATCH",
        data:data
    }).done(function(){
        buildDeviceConfigTable(deviceId);
    });
    // Reset value to blank
    $("#add-new-devCfgItem input").val("");
    // Disable add button
    $("#devcfg-att-btn").prop("disabled", true);
}

$('#device-config-view').on('click', '#deleteDeviceCfgItem', function(event){
    deviceId = $("#device-config-view table").attr("id");
    cfgRow = $(this).parent().parent();
    cfgKey = $(cfgRow).find("td:first").text();
    $.ajax({
        url: "/device/" + deviceId + "/" + cfgKey,
        method: "DELETE"
    }).done(function(){
        buildDeviceConfigTable(deviceId);
    });
});

$('#device-config-view').on('click', '#editDeviceCfgItem', function(event){
    displayButtons(this, ['#doneDeviceCfgItem', '#cancelDeviceCfgItem'], ['#editDeviceCfgItem']);
    var valueElement = $(this).parent().parent().find('td:nth-child(2)');
    valueElement.html('<input type = "text" value="' + valueElement.text() +  '">');
});

$('#device-config-view').on('click', '#doneDeviceCfgItem', function(event){
    displayButtons(this, ['#editDeviceCfgItem'], ['#doneDeviceCfgItem', '#cancelDeviceCfgItem']);
    deviceId = $("#device-config-view table").attr("id");
    deviceName =  $("#device-config-view table").attr("devName");
    deviceParentId =  $("#device-config-view table").attr("parentNodeId");
    cfgRow = $(this).parent().parent();
    cfgKey = $(cfgRow).find("td:first").text();
    cfgValueElem = $(cfgRow).find('td:nth-child(2) input');

    data = '{"Id":' +deviceId + ', ' +
        '"Name": "' + deviceName + '", ' +
        '"ParentNode": ' + deviceParentId + ', ' +
        '"ConfigItems" : ' +
        '[{"Key": "' + cfgKey + '", "Value": "' +  cfgValueElem.val() + '"}]' +
        '}';

    $.ajax({
        url:"/device/" + deviceId,
        method:"PATCH",
        data:data
    }).done(function(){
        cfgValueElem.parent().html(cfgValueElem.val());
    })
    .fail(function(){
        cfgValueElem.parent().html($(cfgValueElem).attr("value"));
    });
});

$('#device-config-view').on('click', '#cancelDeviceCfgItem', function(event){
    displayButtons(this, ['#editDeviceCfgItem'], ['#doneDeviceCfgItem', '#cancelDeviceCfgItem']);
    cfgValueElem = $(this).parent().parent().find('td:nth-child(2) input');
    cfgValueElem.parent().html($(cfgValueElem).attr("value"));
});

/********************* Device Config File view functions *************************/
function buildConfigFileView(deviceId){
    removeAlert();
    //clear off old config File shown in div?
    $("#configFile-content").remove();
    content = "";
    $.get("/device/" + deviceId + "/config", function(data){
        content = data;
    })
    .always(function(){
        $('#device-configFile-view h4').after($('<div id="configFile-content"><textarea style="width: 100%; height: 150px;">' + decodeURIComponent(content) +  '</textarea></div>'));
        setRoleBasedOptions();
    });
}

function onSaveDevCfgFile(){
    removeAlert();
    deviceId = $("#device-config-view table").attr("id");
    deviceName =  $("#device-config-view table").attr("devName");
    content = $("#configFile-content textarea").val();
    msg = $("#msg-devCfgFile input[name=commit-msg]").val();
    data = '{' +
        '"Content": ' + '{' +
        '"Files" : ' +
        '{' +
        '"configFile": "' + encodeURIComponent(content) + '"' +
        '}' +
        ',' +
        '"Object": "' + deviceName + '"' +
        '}' + ', ' + //end of CMContent
        '"Author": ' + '{"Name": "' + $('#UserName').text() + '"}' + ', ' +
        '"Log": ' + '{' +
        '"Message": "' + msg + '"' + '}' +
        '}';

    $.ajax({
        url:"/device/" + deviceId + '/config',
        method:"PATCH",
        contentType: "application/json; charset=utf-8",
        data:data
    })
    .done(function(){
        $("#msg-devCfgFile input").val("");
        $( '<div class="alert alert-success" role="alert">Successfully committed changes to the config file.</div>').insertAfter( "#update-devCfgFile" );
    })
    .fail(function(){
        $("#msg-devCfgFile input").val("");
        $('<div class="alert alert-danger" role="alert">Failed to commit changes to the config file.</div>').insertAfter( "#update-devCfgFile" );
    });
}

function onCancelDevCfgFile(){
    removeAlert();
    $("#msg-devCfgFile input").val("");
    showView("#device-view");
}

/******************* Device Metadata Editing Support ************************/

function showDeviceMetadata(deviceName, deviceId) {
	$("#device-metaFile-view").find("h4").replaceWith('<h4>Metadata for ' + decodeURIComponent(deviceName) + '<small></h4>');
    buildDeviceMetaTableView(deviceId);
    showDeviceMetaFileView();
}

function buildDeviceMetaTableView(deviceId){
    //clear off old config items from the config table
    $("#device-metaFile-view table tbody tr").remove();
    $.get("/device/"+deviceId + "/meta", function(data){
        var metaItems = $.parseJSON(data);
        $("#device-metaFile-view table").attr("id", deviceId);

        for (key in metaItems) {
            row = $('<tr></tr>');
            row.append($('<td>' + key + '</td>'));
            row.append($('<td style="word-wrap:break-word">' + metaItems[key] + '</td>'));
            row.append($('<td class="text-right">'
                                + '<button type="button" id="editDeviceMetaItem" class="btn btn-default" aria-label="Edit"><span class="glyphicon glyphicon-pencil" aria-hidden="true"</span></button>'
                                + '<button type="button" id="doneDeviceMetaItem" class="btn  btn-default hidden" aria-label="Done"><span class="glyphicon glyphicon-ok" aria-hidden="true"</span></button>'
                                + '<button type="button" id="cancelDeviceMetaItem" class="btn  btn-default hidden" aria-label="Cancel"><span class="glyphicon glyphicon-remove" aria-hidden="true"</span></button>'
                                + '<button type="button" id="deleteDeviceMetaItem" class="btn btn-default" aria-label="Delete"><span class="glyphicon glyphicon-trash" aria-hidden="true"</span></button>'
                                +'</td>'));

            $("#device-metaFile-view table tbody").append(row);
        }
    });
}

$('#device-metaFile-view').on('click', '#deleteDeviceMetaItem', function(event){
    cfgRow = $(this).parent().parent();
    cfgKey = $(cfgRow).find("td:first").text();
    cfgValue = $(cfgRow).find('td:nth-child(2)').text();
    deviceId =  $("#device-metaFile-view table").attr("id");

    data = '{"Key":"' + cfgKey + '", "Value" : "' + cfgValue + '"}';
    $.ajax({
        url:"/device/" + deviceId + "/meta",
        method:"DELETE",
        data:data
    }).done(function(){
        buildDeviceMetaTableView(deviceId);
    });
});

$('#device-metaFile-view').on('click', '#editDeviceMetaItem', function(event){
    displayButtons(this, ['#doneDeviceMetaItem', '#cancelDeviceMetaItem'], ['#editDeviceMetaItem']);
    var valueElement = $(this).parent().parent().find('td:nth-child(2)');
    valueElement.html('<input type = "text" value="' + valueElement.text() +  '">');
});

$('#device-metaFile-view').on('click', '#cancelDeviceMetaItem', function(event){
    displayButtons(this, ['#editDeviceMetaItem'], ['#doneDeviceMetaItem', '#cancelDeviceMetaItem']);
    cfgValueElem = $(this).parent().parent().find('td:nth-child(2) input');
    cfgValueElem.parent().html($(cfgValueElem).attr("value"));
});

$('#device-metaFile-view').on('click', '#doneDeviceMetaItem', function(event){
    displayButtons(this, ['#editDeviceMetaItem'], ['#doneDeviceMetaItem', '#cancelDeviceMetaItem']);
    deviceId = $("#device-metaFile-view table").attr("id");
    deviceName =  $("#device-metaFile-view table").attr("devName");
    deviceParentId =  $("#device-metaFile-view table").attr("parentNodeId");
    metaRow = $(this).parent().parent();
    metaKey = $(metaRow).find("td:first").text();
    metaValueElem = $(metaRow).find('td:nth-child(2) input');

    data ='{"Key": "' + metaKey + '", "Value": "' +  metaValueElem.val() + '"}';

    $.ajax({
        url:"/device/" + deviceId + "/meta",
        method:"PATCH",
        data:data
    }).done(function(){
        metaValueElem.parent().html(metaValueElem.val());
    })
    .fail(function(){
        metaValueElem.parent().html($(metaValueElem).attr("value"));
    });
});

/******************* Building dashboard for home view ************************/

// Build tree with nodes/devices
function buildDashboard() {

	// Associative array with nodes as keys and device arrays as values
	var nodeNames = {};
	var nodeDevices = {};
	var deviceNames = {};
	var nodeTreeData = [];

	// Ajax call to get nodes
    $.ajax({
        url: "/node",
        method:"GET"
    }).done(function(data){
        var nodes = $.parseJSON(data);
        for (i=0;i< nodes.length; i++){
        	nodeNames[nodes[i].Id] = nodes[i].Name;
        	nodeDevices[nodes[i].Id] = []
        }

        // Get devices after we have nodes
        $.get("/device", function(data){
            var devices = $.parseJSON(data);

            if (devices != null) {
            	for (i=0;i< devices.length; i++){
            		nodeDevices[devices[i].ParentNode].push(devices[i].Id);
            		deviceNames[devices[i].Id] = devices[i].Name;
            	}
        	}
            var node_keys = Object.keys(nodeNames);
            for (j = 0; j< node_keys.length; j++) {
            	var tmpId = node_keys[j];
            	var tmpName = nodeNames[tmpId];
            	var tmpDevIds = nodeDevices[tmpId];
            	// Need to set up child nodes first so they can be added to parent
            	var childNodes = [];
            	for (k = 0; k < tmpDevIds.length; k++) {
            		var tmpDevId = tmpDevIds[k];
            		var tmpDevName = deviceNames[tmpDevId];
            		// Link to device information
            		var tmpDevHref = '?device-' + tmpDevName + '__' + tmpDevId;
            		// Tree node representing device
            		var tmpChildNode = {text: tmpDevName, href: tmpDevHref, icon: 'glyphicon glyphicon-cog'};
            		childNodes.push(tmpChildNode);
            	}
            	// Link to node information
            	var tmpNodeHref = '?node-' + tmpName + '__' + tmpId;
            	// Tree node representing Node object
            	var tmpParentNode = {text: tmpName, href: tmpNodeHref, nodes: childNodes,
            			state: {expanded: expanded_tree_nodes.indexOf(tmpName) > -1}
            	};
            	nodeTreeData.push(tmpParentNode);
            }

            buildTree(nodeTreeData);
        });
    });
}

// Actually build tree with nodes and devices
function buildTree(nodeTreeData) {

	var $searchableTree = $('#treeview-searchable').treeview({
		data: nodeTreeData,
		enableLinks: true,
		selectedBackColor: "#DDDDDD",
	    selectedColor: "#000000",
	    searchResultColor: "#C78305",

		onNodeExpanded: function(event, data) {
			// Add to array tracking expanded nodes
			expanded_tree_nodes.push(data.text);
		},

		onNodeCollapsed: function(event, data) {
			// Remove from array tracking expanded nodes
			expanded_tree_nodes.splice(data.text,1);
		}
	});

	// Set up search
	var search = function(e) {
		var pattern = $('#input-search').val();
		var options = {
				ignoreCase: !($('#chk-ignore-case').is(':checked')),
				exactMatch: $('#chk-exact-match').is(':checked'),
				revealResults: true
		};
		var results = $searchableTree.treeview('search', [ pattern, options ]);
	}

	// Execute search if any of these things happens
	$('#btn-search').on('click', search);
	$('#input-search').on('keyup', search);
	$('#chk-ignore-case').on('click', search);
	$('#chk-exact-match').on('click', search);

	$('#btn-clear-search').on('click', function (e) {
		$searchableTree.treeview('clearSearch');
		$searchableTree.treeview('collapseAll');
		$('#input-search').val('');
	});

	if ($('#input-search').val() != '') {
		search();
	}
}

// Dashboard panel with reports
function buildDashReports() {
	// Clean up first
	$("#dashboard-reports p").remove();

    // OK to load reports asynchronously -- nothing depends on them
    $.get("/reports", function(data){
        var reports = $.parseJSON(data);
        for (i=0;i< reports.length; i++){
            tmp_val = '<p><span class="glyphicon glyphicon-list-alt" aria-hidden="true"></span> <a href="?report-'+ reports[i] +'">'+ reports[i] + "</p>"
            $("#dashboard-reports").append(tmp_val);
        }
    });
}

// Store times in modifiedTagDash when dashboard is built
// Rebuilding node tree is expensive (slow) and shouldn't happen often
function initializeDashTimer() {
	$.ajax({
        url:"/node",
        method:"HEAD",
    }).done(function(data, textStatus, jqXHR){
    	modifiedTagDash.Nodes = jqXHR.getResponseHeader("Last-Modified");
    });

	$.ajax({
        url:"/device",
        method:"HEAD",
    }).done(function(data, textStatus, jqXHR){
    	modifiedTagDash.Devices = jqXHR.getResponseHeader("Last-Modified");
    });

    $.ajax({
        url:"/reports",
        method:"HEAD"
    }).done(function(data, textStatus, jqXHR){
    	modifiedTagDash.Report = jqXHR.getResponseHeader("X-Pbconf-Report-LastCommitId");
    });
}

/******************* Alarms view in dashboard ************************/

// Checkbox to toggle alarm selection
$( '#home-view').on( 'click', '#alarm-select', function( event, ui ) {
	select_all = this.checked;
	var prob_length = document.getElementById("alarm-table").rows.length;
	for (i = 0; i < prob_length; i++) {
		var row = document.getElementById("alarm-table").rows[i];
		// Check all boxes
		if (select_all) {
			row.cells[0].innerHTML = "<tr><td><div class=\"checkbox\">" +
	          "<label><input type=\"checkbox\" checked=\"checked\" value=\"\">" +
	          alarm_array[i] + "</label></div></td></tr>";
		}
		// Clear all boxes
		else {
			row.cells[0].innerHTML = "<tr><td><div class=\"checkbox\">" +
			"<label><input type=\"checkbox\" value=\"\">" + alarm_array[i] +
			"</label></div></td></tr>";
		}
	}
} );

// Clears selected alarms in alarms table
$('#home-view').on('click', '#clear-alarms', function(event){
	var prob_length = document.getElementById("alarm-table").rows.length;
	// Iterate backwards through the array so offset won't be issue
	for (i = prob_length -1; i > -1; i--) {
		var row = document.getElementById("alarm-table").rows[i];
		var tmpv = row.cells[0].firstChild.firstChild.firstChild;
		if (tmpv.checked == true) {
			// Remove item from array
			alarm_array.splice(i,1);
		}
	}

	// Clean up items in alarms table
	$("#alarm-list form table tr").remove();

	// Repopulate alarms table from array
	for (j = 0; j < alarm_array.length; j++) {
		$("#alarm-list form table").append("<tr><td><div class=\"checkbox\">" +
		        "<label><input type=\"checkbox\" value=\"\">" + alarm_array[j] + "</label></div></td></tr>");
	}

	// Uncheck select box
	$("#alarm-select").prop( "checked", false );
	setAlarmVisibility();
});

// Only show alarm controls if there are alarms to display
function setAlarmVisibility() {
	if (alarm_array.length > 0) {
		$("#alarm-checkbox-div").removeClass('hidden');
		$("#clear-alarms").removeClass('hidden');
	}
	else {
		$("#alarm-checkbox-div").addClass('hidden');
		$("#clear-alarms").addClass('hidden');
	}
}

/******************* Report results table ************************/

// Figure out column types for report results
function checkColTypes(dataStr) {
	lines = dataStr.split("\n");
	// Array of column types for sortable table
	var typesArray = [];
	topLine = lines[0].split("\t");
	// Initialize array with "numeric" value
	for (i = 0; i < topLine.length; i++) {
		typesArray[i] = "numeric";
	}
	
	// If column has NaN value, change type to "alpha"
	for (j = 1; j < lines.length; j++) {
		tmpLine = lines[j].split("\t");
		for (k = 0; k < tmpLine.length; k++) {
			if (isNaN(tmpLine[k])) {
				typesArray[k] = "alpha";
			}
		}
	}
	return typesArray;
}

// Generate table HTML for report results
function reportResultsTable(dataStr) {
	// Get rid of trailing newline so we don't get extra table row
	dataStr = $.trim(dataStr);
	beginTable = "<table class=\"sortable-theme-bootstrap\" data-sortable width=\"100%\" id=\"report-res-table\">";
	beginHead = "\n\t<thead>\n\t\t<tr>\n";
	endHead = "\t\t</tr>\n\t</thead>\n";
	endTable = "</table>";
	
	// Table headers
	lines = dataStr.split("\n");
	headerCols = lines[0].split("\t");
	colTypes = checkColTypes(dataStr);
	// Divide into equal columns
	colWidth = Math.floor(100/headerCols.length);
	headContent = "";
	for (i = 0; i < headerCols.length; i++) {
		headContent = headContent + "\t\t\t<th data-sortable-type=\"" + colTypes[i] + "\" width=\"" + colWidth.toString() +"%\">" + headerCols[i] + "</th>\n";
	}
	
	// Table body
	bodyContent = "\n\t<tbody>\n";
	for (j = 1; j < lines.length; j++) {
		bodyContent =  bodyContent + "\t\t<tr>\n";
		bodyArr = lines[j].split("\t");
		for (k = 0; k < bodyArr.length; k++) {
			bodyContent = bodyContent + "\t\t\t<td>" + bodyArr[k] + "</td>\n";
		}
		bodyContent = bodyContent + "\t\t</tr>\n";
	}
	bodyContent = bodyContent +  "\t</tbody>\n";
	var headVal = beginHead + headContent + endHead;
	return beginTable + headVal + bodyContent + endTable;
}

/******** Capture return/enter keystrokes ************/

$("#new-device-name").keyup(function(event){
    if(event.keyCode == 13){
        $("#new-device-btn").click();
    }
});

$("#new-policy-name").keyup(function(event){
    if(event.keyCode == 13){
        $("#new-policy-btn").click();
    }
});

$("#new-report-name").keyup(function(event){
    if(event.keyCode == 13){
        $("#new-report-btn").click();
    }
});

$("#new-config-att").keyup(function(event){
    if(event.keyCode == 13){
        $("#new-config-btn").click();
    }
});

$("#devcfg-att-val").keyup(function(event){
    if(event.keyCode == 13){
        $("#devcfg-att-btn").click();
    }
});

$("#devmeta-att-val").keyup(function(event){
    if(event.keyCode == 13){
        $("#devmeta-att-btn").click();
    }
});

$("#policyfile-change-msg").keyup(function(event){
	var btn_disabled = $("#policyfile-change-btn").is(":disabled");
	if (!btn_disabled) {
		if(event.keyCode == 13){
			$("#policyfile-change-btn").click();
		}
	}
});

$("#dev-cfg-chng").keyup(function(event){
    if(event.keyCode == 13){
        $("#dev-cfg-save-btn").click();
    }
});

$("#report-commit-msg").keyup(function(event){
	var btn_disabled = $("#rpt-button-save").is(":disabled");
	if (!btn_disabled) {
		if(event.keyCode == 13){
			$("#rpt-button-save").click();
		}
	}
});

/******** Enable/disable "Add" buttons ************/

// New node configuration item
$("#new-config-att").on('input',function(){
	if ($("#new-config-att").val()) {
		$("#new-config-btn").prop("disabled", false);
	}
	else {
		$("#new-config-btn").prop("disabled", true);
	}
});

// New device name
$("#new-device-name").on('input',function(){
	if ($("#new-device-name").val()) {
		$("#new-device-btn").prop("disabled", false);
	}
	else {
		$("#new-device-btn").prop("disabled", true);
	}
});

// New device attribute name
$("#devcfg-att-name").on('input',function(){
	if ($("#devcfg-att-name").val()) {
		$("#devcfg-att-btn").prop("disabled", false);
	}
	else {
		$("#devcfg-att-btn").prop("disabled", true);
	}
});

// New device metadata attribute name
$("#devmeta-att-name").on('input',function(){
	if ($("#devmeta-att-name").val()) {
		$("#devmeta-att-btn").prop("disabled", false);
	}
	else {
		$("#devmeta-att-btn").prop("disabled", true);
	}
});

// New policy name
$("#new-policy-name").on('input',function(){
	if ($("#new-policy-name").val()) {
		$("#new-policy-btn").prop("disabled", false);
	}
	else {
		$("#new-policy-btn").prop("disabled", true);
	}
});

// New report name
$("#new-report-name").on('input',function(){
	if ($("#new-report-name").val()) {
		$("#new-report-btn").prop("disabled", false);
	}
	else {
		$("#new-report-btn").prop("disabled", true);
	}
});

