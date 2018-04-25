/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import org.apache.log4j.Logger;
import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLOntology;

/**
 * Used to parse the JSON object sent from PBConf and process it in the ontology
 * @author JD
 */
public class OntologyParser {
    static Logger logger = Logger.getLogger(OntologyParser.class.getName().replaceFirst(".+\\.", ""));
    //testMode var is used for unit tests, to make sure we don't keep content in ontology on exit
    private boolean testMode = false;

    /**
     * Default constructor to start the ontology parser, which processes input from the ontology client.
     * This can be JSON objects, or an exit command to close the connection
     */
    public OntologyParser() {
        //logger.info("Parser started");
    }
       
    /**
     * Run a "config" command using the new buffer / translation setup
     * This sort of invalidates a bunch of the other code above this but leaving it for now
     * Will need to move some of this outside the for loop that calls this fn
     * @param ontStr
     * @param gizerStr
     * @param op
     * @param key
     * @param val
     * @return 
     */
    private Boolean runConfigCommand(OWLOntology partialOnt, String device, String ontologizer, String Op, String Key, String Val, String Svc) {
        if (!Ontologizer.isValidOntologizer(ontologizer)) {
            return false;
        }
        
        Ontologizer gizer = Ontologizer.getOntologizer(ontologizer); 
        boolean state = gizer.setProperty(partialOnt, device, Op, Key, Val, Svc);
        
        if (state == false) {
            return false;
        } else {
            return Ontology.instance().isConstistent();
        }
    }
   
    /**
     * Process a list of property arrays, return consistency
     * @param individual
     * @param ontologizer
     * @param deviceProperties
     * @param partialOnt 
     * @return  
     */
    public String processConfigCommands(String individual, String ontologizer, JSONArray deviceProperties, OWLOntology partialOnt) {
        //Loop, parse each property, execute and validate consistency.
        boolean outerValid = true;
        boolean innerValid = true;
        String command = "";
        for (int i = 0; i < deviceProperties.length(); i++) {
            if (outerValid == false) {
                continue;
            }
            JSONObject tempObj = deviceProperties.getJSONObject(i);
            String Op = tempObj.getString("Op");
            String Key = tempObj.getString("Key");
            String Val = tempObj.getString("Val");
            String Svc = "";
            //Only use case for when Svc value is set
            if ("service_option".equals(Op.toLowerCase()) && tempObj.has("Svc")) {
                Svc = tempObj.getString("Svc");
            }
            
            innerValid = runConfigCommand(partialOnt, individual, ontologizer, Op, Key, Val, Svc);
            if (innerValid == false) {
                logger.info("Validation was failed, explanation : ");
                logger.info(Ontology.instance().getFriendlyExplanations(Ontology.instance().getExplanations()));
                
                outerValid = false;
                command += individual + ", ";
                command += ontologizer + ", ";
                command += Op + ", ";
                command += Key + ", ";
                if (!"".equals(Svc)) {
                    command = "Command failure (individual, ontologizer, Op, key, Val, Svc) : " + command; 
                    command += Val + ", ";
                    command += Svc;
                } else {
                    command = "Command failure (individual, ontologizer, Op, key, Val) : " + command; 
                    command += Val;
                }
            }
        }
        
        if (outerValid == true) {
            return "TRUE";
        } else {
            return "FALSE:" + command;
        }
    }
    
    /**
     * Parse and process an array of configuration commands about a single individual / device
     * This is now the only format we'll be accepting: A set of configurations targeting a single individual.
     * @param jObj
     * @return 
     */
    public String parseAndProcessSingleConfigCommand(JSONObject jObj) {
        logger.info("Parsing and processing single config command");
        String status = "";
        String explanation = "\"\"";
        
        Ontology.instance().reloadBaseOntologies(true, true);
        Ontology.instance().addTemporaryConfigurationOntology();
        
        String ontologizer = jObj.getString("ontologizer");
        String individual = jObj.getString("individual");
        JSONArray innerProperties = jObj.getJSONArray("properties");

        OWLOntology partial = Ontology.instance().getOntology(Ontology.instance().getConfig().get("partialConfigOntology"));
        String result = processConfigCommands(individual, ontologizer, innerProperties, partial);
        if (result.startsWith("TRUE")) {
            if (!Ontology.instance().isConstistent() || !ClosedWorld.reason(testMode)) {
                status = "INVALID";
                explanation = Ontology.instance().getFriendlyExplanations(Ontology.instance().getExplanations());
                //No need to save since we just did that, we can simply reset
                Ontology.instance().reloadBaseOntologies(false, false);
            } else {
                status = "VALID";
                explanation = "";
                //Now that we know this was a valid ocnfiguration, we can clear that individual from the config ontology
                //before adding it in by copying it over from the temporary one.
                Ontology.instance().addConfigurationOntology();
                Ontology.instance().clearIndividual(Ontology.instance().getConfig().getConfigPrefix(), individual);
                Ontology.instance().copyTemporaryToConfiguration();
                Ontology.instance().reloadBaseOntologies(true, true);
            } 
        } else {
            status = "INVALID";
            explanation = result.split("FALSE:")[1];
            explanation = "\"" + explanation + "\"";
            Ontology.instance().reloadBaseOntologies(false, false);
        }
        
        if ("".equals(explanation)) { explanation = "\"\""; }
        return "{\"status\":\"" + status + "\",\"explanation\":" + explanation + "}";
    }
    
    /**
     * Parse and process a configuration command
     * Configuration commands contain (with examples)
     * {
     *   ontology: config
     *   properties: [
     *     ontologizer: (SEL421 || sel421 || linux)
     *     individual: sel421FAKEDEVICEA || ...
     *     properties: []
     *   ]
     * }
     * @param jObj
     * @return 
     */
    public String parseAndProcessConfigCommand(JSONObject jObj) {
        return parseAndProcessSingleConfigCommand(jObj);
    }
    
    /**
     * Add a policy axiom to the policy ontology
     * policies don't use a specific ontologizer since they should operate
     * independently of that.
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @return 
     */
    private void addPolicyAxiom(String target, String subject, String predicate, Object object) {
        OWLOntology ont = Ontology.instance().getOntology(Ontology.instance().getConfig().getPolicyPrefix());

        //Fallback for when target is set, but subject isn't.
        if (!"".equals(target) && "".equals(subject)) {
            subject = target;
        }
        
        //Determine if we need to use the ontology string parser to parse 
        //equivalentTo/disjointWith statements, or use the simpler ontology policy engine
        if (predicate.toLowerCase().equals("disjointwith") || predicate.toLowerCase().equals("equivalentto")) {
            OntologyPolicyStringEngine opse = new OntologyPolicyStringEngine();
            OWLAxiom ax = opse.constructAxiomFromString(subject, predicate, (String)object);
            if (ax != null) {
                opse.applyAxiom(ax);
            }
        } else {
            OntologyPolicyEngine ope = new OntologyPolicyEngine();
            ope.addAxiom(ont, target, subject, predicate, object, testMode);
        }
    }
    
    /**
     * Process policy commands sent from PBConf
     * We've already removed all elements from the policy ontology in the parent function
     * For this, we just loop through all axioms and apply them
     * @param target
     * @param axioms
     */
    public void processPolicyAxioms(String target, JSONArray axioms) {
        //Loop, parse each property, execute and validate consistency.
        for (int i = 0; i < axioms.length(); i++) {
            JSONObject tempObj = axioms.getJSONObject(i);
            String subject = tempObj.getString("s");
            String predicate = tempObj.getString("p");
            Object object = tempObj.get("o");
            addPolicyAxiom(target, subject, predicate, object);
        }
    }
    
    /**
     * Parse and process a policy command
     * Policy commands work independent of device type, so that's not passed
     * Policy commands contain
     * {
     *   ontology: policy
     *   data: []
     * }
     * @param jObj
     * @return 
     */
    public String parseAndProcessPolicyCommand(JSONObject jObj) {
        String status = "";
        String explanation = "";
        OntologyConfig cfg = Ontology.instance().getConfig();
        Ontology.instance().reloadBaseOntologies(true, true);
        //Ontology.instance().addConfigurationOntology();
        Ontology.instance().replaceOntology(cfg.getPolicyPrefix());
        JSONArray dataArr = jObj.getJSONArray("data");
        for (int i = 0; i < dataArr.length(); i++) {
            JSONArray subDataArr = dataArr.getJSONArray(i);
            for (int j = 0; j < subDataArr.length(); j++) {
                JSONObject dataObj = subDataArr.getJSONObject(j);
                String target = dataObj.getString("Class");
                JSONArray axioms = dataObj.getJSONArray("Axioms");
                processPolicyAxioms(target, axioms);
            }
        }
        //We're no longer validating on new policy, just allowing it through. So now we just save and reset
        Ontology.instance().reloadBaseOntologies(true, true);
        return "{\"status\":\"VALID\",\"explanation\":\"\"}";
    }
    
    /**
     * Parse and process a JSON command from PBConf
     * @param jObj
     * @return 
     */
    public String processJSONCommand(JSONObject jObj) {
        String commandResult;
        
        //Right now, we're only accepting policy and config ontology selection
        //Outside of that will need to be configured
        String ont = jObj.getString("ontology").toLowerCase(); 
        switch (ont) {
            case "policy":
                commandResult = parseAndProcessPolicyCommand(jObj);
                break;
            case "config":
                commandResult = parseAndProcessConfigCommand(jObj);
                break;
            default:
                commandResult = "{\"status\":\"INVALID\",\"explanation\":\"Invalid Ontology Choice\"}";
                break;
        }
        
        //commandResult = commandResult.replace("\"", "");
        logger.info("=================Issued Command=================");
        logger.info(jObj.toString());
        logger.info("=================Command Result=================");
        logger.info(commandResult);
        return commandResult;
    }
    
    /**
     * Process a string command, now only accepting exit or close
     * @param input
     * @return 
     */
    public String processStrCommand(String input) {
        String output = "";
        
        if (isExitCommand(input)) {
            logger.info("Proccessing exit command");
            output = "";
        } else if (isValidateCommand(input)) {
            return Ontology.instance().validate("standard");
        } else if (isResetCommand(input)) {
            logger.info("An ontlogy reset was issued, replacing");
            Ontology.instance().reloadForTests();           
            output = "";
        } else {
            output = "Invalid command";
        }
        
        return output;
    }
    
    /**
     * Check to see if PBConf requested a complete ontology validation.
     * This will go through all configurations and find issues with each for the current policy.
     * @param input
     * @return 
     */
    public boolean isValidateCommand(String input) {
        String inputs[] = null;
        
        if (input.contains(" ")) {
            inputs = input.split("\\s+");
        } else {
            inputs = new String[1];
            inputs[0] = input;
        }
        
        return "validate".equals(inputs[0].toLowerCase());
    }
    
     /**
     * Check to see if PBConf requested a complete ontology reset.
     * This will reset the policy and configuration ontologies
     * @param input
     * @return 
     */
    public boolean isResetCommand(String input) {
        String inputs[] = null;
        
        if (input.contains(" ")) {
            inputs = input.split("\\s+");
        } else {
            inputs = new String[1];
            inputs[0] = input;
        }
        
        return "reset".equals(inputs[0].toLowerCase());
    }
    
    /**
     * Helper function to determine if we're ending a connection based on input
     * @param input
     * @return 
     */
    public Boolean isExitCommand(String input) {
        String inputs[] = null;
        
        if (input.contains(" ")) {
            inputs = input.split("\\s+");
        } else {
            inputs = new String[1];
            inputs[0] = input;
        }
        
        return "exit".equals(inputs[0]) || "close".equals(inputs[0]);
    }
    
    /**
     * Process JSON input
     * @param input
     * @return 
     */
    public JSONObject processInput(String input) {
        JSONObject resultObj = new JSONObject();
        //Try to parse out a valid json object.  If it's valid, we can 
        //pull the ontology info and determine if we're executing a config
        //command, or a policy command.  If it fails to parse, try running a
        //standard client command option.
        try {
            JSONObject jObj = new JSONObject(input);
            resultObj.put("output", processJSONCommand(jObj));
            resultObj.put("exiting", false);
            return resultObj;
        } catch (JSONException jException) {        
            //logger.info("Parsing exit command");
            //Please note - you will get an exception here as part of the pbontology
            //process.  This is expected because after you pass a result, it tells
            //the client connection to exit, which isn't json, it's just a command.
        
            String output = processStrCommand(input);
            if (isValidateCommand(input)) {
                logger.info("VALIDATION RESPONSE");
                logger.info(output);
            }
            
            resultObj.put("output", output);
            resultObj.put("exiting", isExitCommand(input));
                       
            return resultObj;
        }
    }
    
    /**
     * Wrapper for the processInput function for use with unit tests.
     * Simply sets testMode to true for the instance of OntologyParser
     * This is used later on and doesn't actually write to in memory ontology
     * @param input
     * @return 
     */
    public JSONObject processTestInput(String input) {
        testMode = true;
        return processInput(input);
    }
}
