/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.PBConfParserTest.logger;
import java.io.FileNotFoundException;
import java.util.ArrayList;
import java.util.List;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.AfterClass;
import org.junit.Before;
import org.junit.BeforeClass;
import org.semanticweb.owlapi.model.OWLException;
import static org.junit.Assert.*;
import org.junit.Test;
import org.semanticweb.owlapi.model.OWLOntology;

/**
 * Test the equivalentTo statement for string handling
 * @author Joe DiGiovanna
 */
public class ReleaseTest {
    static Logger logger = Logger.getLogger(OntologyPolicyStatementTest.class.getName().replaceFirst(".+\\.",""));

    /**
     * Setup basic configurator
     */
    @BeforeClass 
    public static void once() {
        // setup logger
        BasicConfigurator.configure();
    }
    
    /**
     * After a test class, we will reset policy and configuration
     */
    @AfterClass
    public static void afterClass() {
        //Clean everything up      
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        try {
            Ontology.instance().initialize(cfg);
        } catch (FileNotFoundException ex) {
            logger.error("base ontology file", ex);
        } catch (OWLException ex) {
            logger.error("OWLException", ex);
        }
  
        Ontology.instance().reloadForTests();
    }
    
    /**
     * Setup ontology and some example configuration data
     */
    @Before
    public void setUp() {       
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        try {
            Ontology.instance().initialize(cfg);
        } catch (FileNotFoundException ex) {
            logger.error("base ontology file", ex);
        } catch (OWLException ex) {
            logger.error("OWLException", ex);
        }

        Ontology.instance().reloadBaseOntologies(false, false);
        Ontology.instance().addConfigurationOntology();
        Ontology.instance().replaceOntology("config");     
        Ontology.instance().replaceOntology("policy");     
        Ontology.instance().reloadBaseOntologies(true, true);
        Ontology.instance().addConfigurationOntology();
    }

    /**
     * Use the processInput function in OntologyParser to process the JSON data
     * This can accept both config and policy statements
     *
     * @param jsonStr
     * @param expectedResult
     * @return
     */
    public boolean operateAndCompare(String jsonStr, boolean expectedResult) {
        OntologyParser parser = new OntologyParser();
        JSONObject result = parser.processTestInput(jsonStr);
        String outputStr = result.getString("output");

        JSONObject output = null;
        String status = "";

        try {
            output = new JSONObject(outputStr);
            status = output.getString("status");
        } catch (JSONException ex) {
            logger.info(ex.toString());
        }

        if ("VALID".equals(status) && expectedResult == true) {
            return true;
        }

        return ("INVALID".equals(status) || "ERROR".equals(status)) && expectedResult == false;
    }
    
    /**
     * Set a valid disjoint statement that implies we're disjoint with something calle FAKECLASS
     * @throws Exception 
     */
    @Test
    public void sel421_statement_test_0() throws Exception {
       /**
        * This is so that we can add policy on existing configuration to test it.
        */
        List dataArray = new ArrayList();
        
        String[] axioms = new String[2];
        axioms[0] = "{\"s\":\"password.level2\",\"p\":\"min-length\",\"o\":\"4\"}";
        axioms[1] = "{\"s\":\"password.level2\",\"p\":\"max-length\",\"o\":\"16\"}";
        
        dataArray.add(DataGenerator.buildJSONPolicyRequestStr("LINUX", axioms));
        String policyStatement = DataGenerator.buildFullJSONPolicyRequestStr("LINUX", dataArray);
        
        //This should be considered valid since users are only set on a LINUX box, not an SEL421
        assertTrue(operateAndCompare(policyStatement, true));
        
        List configArray = new ArrayList();

        configArray.add(DataGenerator.buildJSONRequestStr("variable", "type", "LINUX", ""));
        configArray.add(DataGenerator.buildJSONRequestStr("password", "level2", "testPass", ""));      
        String configStatement = DataGenerator.buildFullJSONRequestStr("LINUX", "linuxa", configArray);
        
        assertTrue(operateAndCompare(configStatement, true));

        List badConfigArray = new ArrayList();

        badConfigArray.add(DataGenerator.buildJSONRequestStr("variable", "type", "LINUX", ""));
        badConfigArray.add(DataGenerator.buildJSONRequestStr("password", "level2", "bad", ""));      
        String badConfigStatement = DataGenerator.buildFullJSONRequestStr("LINUX", "linuxa", badConfigArray);
        
        logger.info(badConfigStatement);
        
        assertTrue(operateAndCompare(badConfigStatement, false));
        
        Ontology.instance().reloadBaseOntologies(false, false);
    }
    
    @Test
    public void basicPasswordPolicyStringTest() throws Exception {  
	String policy = "{\"ontology\":\"policy\",\"data\":[[{\"Class\":\"LINUX\",\"Axioms\":[{\"s\":\"password.level2\",\"p\":\"min-length\",\"o\":\"8\"},{\"s\":\"password.level2\",\"p\":\"max-length\",\"o\":\"16\"}]}]]}";
	String goodConfiguration = "{\"individual\":\"linuxa\",\"ontologizer\":\"LINUX\",\"ontology\":\"config\",\"properties\":[{\"Val\":\"LINUX\",\"Op\":\"variable\",\"Svc\":\"\",\"Key\":\"type\"},{\"Val\":\"validPass\",\"Op\":\"password\",\"Svc\":\"\",\"Key\":\"level2\"}]}";
	String badConfiguration = "{\"individual\":\"linuxa\",\"ontologizer\":\"LINUX\",\"ontology\":\"config\",\"properties\":[{\"Val\":\"LINUX\",\"Op\":\"variable\",\"Svc\":\"\",\"Key\":\"type\"},{\"Val\":\"badPass\",\"Op\":\"password\",\"Svc\":\"\",\"Key\":\"level2\"}]}";
        
        Ontology.instance().addConfigurationOntology();
        
        assertTrue(operateAndCompare(policy, true));
        OWLOntology polOnt = Ontology.instance().getOntology("policy");
                 
        assertTrue(operateAndCompare(goodConfiguration, true)); 
        Ontology.instance().addConfigurationOntology(); 
        OWLOntology cfgOnt = Ontology.instance().getOntology("config");
        
        Ontology.instance().addConfigurationOntology();       
        assertTrue(operateAndCompare(badConfiguration, false));
    }
}