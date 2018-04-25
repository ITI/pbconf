/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
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

/**
 * New configuration tests to verify that the new methods are working
 * @author Joe
 */
public class OntologyNewRulesTest {
    static Logger logger = Logger.getLogger(OntologyNewRulesTest.class.getName().replaceFirst(".+\\.",""));
    
    private static boolean setUpIsDone = false;
    
    /**
     * Setup basic configuration
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
    public static void after() {
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
     * load ontologies and setup default test data
     */
    @Before
    public void setUp() {
        if (setUpIsDone) {
            return;
        }
        
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

        setUpIsDone = true;
    }
    
    /**
     * Build out a full configuration statement so we have real configuration data
     * @param ontologizer
     * @param ind
     * @param dataArray
     * @return 
     */
    public String buildFullJSONRequestStr(String ontologizer, String ind, List dataArray) {
        String jsonStr = "{\"ontology\":\"config\", \"ontologizer\":\"" + ontologizer + "\",\"individual\":\"" + ind + "\",\"properties\":[";
        
        for (int i = 0; i < dataArray.size(); i++) {
            jsonStr = jsonStr.concat((String) dataArray.get(i));
            if (i != dataArray.size() - 1) {
                jsonStr = jsonStr.concat(",");
            }
        }
        
        jsonStr += "]}";
        
        return jsonStr;
    }
    
    /**
     * Construct a JSON request for parsing, for a configuration operation
     * @param Op
     * @param Key
     * @param Val
     * @param Svc
     * @return 
     */
    public String buildJSONRequestStr(String Op, String Key, String Val, String Svc) {
        String jsonStr = "{\"Op\":\"" + Op + "\",\"Key\":\"" + Key + "\",\"Val\":\"" + Val + "\", \"Svc\":\"" + Svc + "\"}";

        return jsonStr;
    }
    
    /**
     * Quick save by prefix
     * @param prefix 
     */
    public void save(String prefix) {
        try {
            Ontology.instance().saveOntologyByPrefix(prefix);
        } catch (OWLException | FileNotFoundException ex) {
            logger.error(ex.toString());
        }
    }
    
    /**
     * Use the processInput function in OntologyParser to process the JSON data
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
     * This will set a level 1 password of "testPass" for sel421FAKEDEVICEB
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_password_0() throws Exception {
        Ontology.instance().reloadForTests();
        
        List dataArray = new ArrayList();
        dataArray.add(buildJSONRequestStr("type", "", "SEL421", ""));
        dataArray.add(buildJSONRequestStr("password", "hasLvl1Pwd", "testPass", ""));
        String dataStatement = buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", dataArray);
        assertTrue(operateAndCompare(dataStatement, true));
        
        assertTrue("".equals(Ontology.instance().validate("test")));
        
        assertTrue(operateAndCompare(DataGenerator.getPasswordPolicy(), true));
        
        boolean curState = ("".equals(Ontology.instance().validate("test")));
        
        assertTrue(!curState);
        
        List dataArrayB = new ArrayList();
        dataArrayB.add(buildJSONRequestStr("type", "", "SEL421", ""));
        dataArrayB.add(buildJSONRequestStr("password", "hasLvl1Pwd", "testpass", ""));
        String dataStatementB = buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", dataArrayB);
        assertTrue(operateAndCompare(dataStatementB, true));
        
        assertTrue("".equals(Ontology.instance().validate("test")));
    }
    
    /**
     * This will set a service DNP3 to the on state
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_service_0() throws Exception {
        List dataArray = new ArrayList();
        dataArray.add(buildJSONRequestStr("service", "hasDNP3Stt", "on", ""));
        String dataStatement = buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", dataArray);
        assertTrue(operateAndCompare(dataStatement, true));
    }
    
    /**
     * This will create a new individual called 'zebra', and set hasDNPStt to that
     * It will error because that is an invalid range.
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_service_1() throws Exception {
        List dataArray = new ArrayList();
        dataArray.add(buildJSONRequestStr("service", "hasDNP3Stt", "zebra", ""));
        String dataStatement = buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", dataArray);
        assertTrue(operateAndCompare(dataStatement, false));
    }
   
}