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
 * Test the equivalentTo statement for string handling
 * @author Joe DiGiovanna
 */
public class OntologyPolicyStatementTest {
    static Logger logger = Logger.getLogger(OntologyPolicyStatementTest.class.getName().replaceFirst(".+\\.",""));
    
    private static boolean setUpIsDone = false;
    
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
     * Setup ontology and some example configuration data
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

        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        
        setUpIsDone = true;
    }
    
    /**
     * Quick save by prefix
     *
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
     * This can accept both config and policy statements
     *
     * @param jsonStr
     * @param expectedResult
     * @return
     */
    public boolean operateAndCompare(String jsonStr, boolean expectedResult) {
        OntologyParser parser = new OntologyParser();
        JSONObject result = parser.processInput(jsonStr);
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
        
        String[] axioms = new String[1];
        axioms[0] = "{\"s\":\"SEL421\",\"p\":\"disjointWith\",\"o\":\"FAKECLASS\"}";
        
        dataArray.add(DataGenerator.buildJSONPolicyRequestStr("SEL421", axioms));
        String dataStatement = DataGenerator.buildFullJSONPolicyRequestStr("SEL421", dataArray);
        
        //This should be considered valid since users are only set on a LINUX box, not an SEL421
        assertTrue(operateAndCompare(dataStatement, true));
        assertTrue("".equals(Ontology.instance().validate("test")));
    }
    
    /**
     * Set a valid disjoint statement that implies we're disjoint with something called FAKECLASS
     * Then set an invalid statement saying we're equivalent to that (should be inconsistent after that)
     * @throws Exception 
     */
    @Test
    public void sel421_statement_test_1() throws Exception {
       /**
        * This is so that we can add policy on existing configuration to test it.
        */
        List dataArray = new ArrayList();
        
        String[] axioms = new String[2];
        axioms[0] = "{\"s\":\"SEL421\",\"p\":\"disjointWith\",\"o\":\"FAKECLASS\"}";
        axioms[1] = "{\"s\":\"SEL421\",\"p\":\"equivalentTo\",\"o\":\"FAKECLASS\"}";
        
        dataArray.add(DataGenerator.buildJSONPolicyRequestStr("SEL421", axioms));
        String dataStatement = DataGenerator.buildFullJSONPolicyRequestStr("SEL421", dataArray);
        
        assertTrue(operateAndCompare(dataStatement, true));
        assertTrue(!"".equals(Ontology.instance().validate("test")));
    }
    
    /**
     * Set several valid disjoint statements that use some advanced formatting
     * @throws Exception 
     */
    @Test
    public void sel421_statement_test_2() throws Exception {
        /**
        * This is so that we can add policy on existing configuration to test it.
        */
        List dataArray = new ArrayList();
        
        String[] axioms = new String[1];
        axioms[0] = "{\"s\":\"SEL421\",\"p\":\"disjointWith\",\"o\":\"(authDNP $some Person) $or (authFTP $some Person $and (authFTP $some IPAddress)) $or (authTelnet $some Person) $or (hasLvl2Pwd $value test)\"}";
               
        dataArray.add(DataGenerator.buildJSONPolicyRequestStr("SEL421", axioms));
        String dataStatement = DataGenerator.buildFullJSONPolicyRequestStr("SEL421", dataArray);
        
        //This should be considered valid since users are only set on a LINUX box, not an SEL421
        assertTrue(operateAndCompare(dataStatement, true));
        assertTrue("".equals(Ontology.instance().validate("test")));
    }
    
    /**
     * Set several valid disjoint statements that use some advanced formatting
     * Make sure class isn't set, but should be derived from the target
     * @throws Exception 
     */
    @Test
    public void sel421_statement_test_3() throws Exception {
        /**
        * This is so that we can add policy on existing configuration to test it.
        */
        List dataArray = new ArrayList();
        
        String[] axioms = new String[1];
        axioms[0] = "{\"s\":\"\",\"p\":\"disjointWith\",\"o\":\"(authDNP $some Person) $or (authFTP $some Person $and (authFTP $some IPAddress)) $or (authTelnet $some Person) $or (hasLvl2Pwd $value test)\"}";
               
        dataArray.add(DataGenerator.buildJSONPolicyRequestStr("SEL421", axioms));
        String dataStatement = DataGenerator.buildFullJSONPolicyRequestStr("SEL421", dataArray);
        
        //This should be considered valid since users are only set on a LINUX box, not an SEL421
        assertTrue(operateAndCompare(dataStatement, true));
        assertTrue("".equals(Ontology.instance().validate("test")));
    }
}