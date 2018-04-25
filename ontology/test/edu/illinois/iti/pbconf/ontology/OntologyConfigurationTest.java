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
 *
 * @author anderson
 */
public class OntologyConfigurationTest {
    static Logger logger = Logger.getLogger(OntologyConfigurationTest.class.getName().replaceFirst(".+\\.",""));
    
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

        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }

        setUpIsDone = true;
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
     * This will validate because alex and bob don't exist in the configuration ontology
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_specific_valid() throws Exception {
        Ontology.instance().prepareForTesting();
        
        Ontologizer gizer = Ontologizer.getOntologizer("sel421");
        logger.info(gizer.getClass().toString());
        boolean test = gizer.isDeviceSpecificValid("service_option", "authUser", "Alex Bob", "hasTelnetStt");
        logger.info("Test result : " + test);
        assertTrue(test == true);
    }
    
    /**
     * This will won't validate because Joe and Andy do exist in the ontology
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_specific_valid_b() throws Exception {
        Ontology.instance().prepareForTesting();
        
        Ontologizer gizer = Ontologizer.getOntologizer("sel421");
        logger.info(gizer.getClass().toString());
        boolean test = gizer.isDeviceSpecificValid("service_option", "authUser", "Joe Andy", "hasTelnetStt");
        logger.info("Test result : " + test);
        assertTrue(test == false);
    }
    
    /**
     * This will set a level 1 password of "testPass" for sel421FAKEDEVICEB
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_password_0() throws Exception {
        Ontology.instance().prepareForTesting();    
        List configList = new ArrayList();
        configList.add(DataGenerator.buildJSONRequestStr("password", "hasLvl1Pwd", "testPass", ""));
        String request = DataGenerator.buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", configList);
        assertTrue(operateAndCompare(request, true));
    }
    
    /**
     * This will set a service dnp to the on state
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_service_0() throws Exception {
        Ontology.instance().prepareForTesting();    
        List configList = new ArrayList();
        configList.add(DataGenerator.buildJSONRequestStr("service", "hasDNP3Stt", "on", ""));
        String request = DataGenerator.buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", configList);
        assertTrue(operateAndCompare(request, true));
    }
    
    /**
     * This will create a new individual called 'zebra', and set hasDNPStt to that
     * It will error because that is an invalid range.
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_service_1() throws Exception {       
        Ontology.instance().prepareForTesting();    
        List configList = new ArrayList();
        configList.add(DataGenerator.buildJSONRequestStr("service", "hasDNP3Stt", "zebra", ""));
        String request = DataGenerator.buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", configList);
        assertTrue(operateAndCompare(request, false));
    }
    
    /**
     * This will create a new individual called 'zebra', and set hasDNPStt to that
     * It will error because that is an invalid range.
     * @throws Exception 
     */
    @Test
    public void sel421FAKEDEVICEB_configuration_service_2() throws Exception {      
        Ontology.instance().prepareForTesting();    
        List configList = new ArrayList();
        configList.add(DataGenerator.buildJSONRequestStr("service", "hasPort5", "sel421FAKEDEVICEB_port5", ""));
        String request = DataGenerator.buildFullJSONRequestStr("SEL421", "sel421FAKEDEVICEB", configList);
        assertTrue(operateAndCompare(request, true));
    }
    
    /**
     * Setup Linux device test configuration
     * @throws Exception
     */
    @Test
    public void linuxbox_configuration_a() throws Exception {
        Ontology.instance().prepareForTesting();    
        String request = "";
        List peopleList = new ArrayList();
        
        request = "";
        peopleList = new ArrayList();
        peopleList.add(DataGenerator.buildJSONRequestStr("type", "", "Person", ""));
        request = DataGenerator.buildFullJSONRequestStr("LINUX", "Joe", peopleList);
        assertTrue(operateAndCompare(request, true));
        
        request = "";
        peopleList = new ArrayList();
        peopleList.add(DataGenerator.buildJSONRequestStr("type", "", "Person", ""));
        request = DataGenerator.buildFullJSONRequestStr("LINUX", "Adrian", peopleList);
        assertTrue(operateAndCompare(request, true));
        
        request = "";
        peopleList = new ArrayList();
        peopleList.add(DataGenerator.buildJSONRequestStr("type", "", "Person", ""));
        request = DataGenerator.buildFullJSONRequestStr("LINUX", "Andy", peopleList);
        assertTrue(operateAndCompare(request, true));
        
        request = "";
        peopleList = new ArrayList();
        peopleList.add(DataGenerator.buildJSONRequestStr("type", "", "Person", ""));
        request = DataGenerator.buildFullJSONRequestStr("LINUX", "Jeremy", peopleList);
        assertTrue(operateAndCompare(request, true));
        
        List configList = new ArrayList();
        configList.add(DataGenerator.buildJSONRequestStr("variable", "type", "LINUX", ""));
        configList.add(DataGenerator.buildJSONRequestStr("service", "authusers", "on", ""));
        configList.add(DataGenerator.buildJSONRequestStr("service", "anonymousftp", "on", ""));
        configList.add(DataGenerator.buildJSONRequestStr("service_option", "authusers", "Joe Andy Adrian Jeremy", "telnet"));
        configList.add(DataGenerator.buildJSONRequestStr("service_option", "authusers", "Joe Andy Adrian Jeremy", "ftp"));
        configList.add(DataGenerator.buildJSONRequestStr("service_option", "authusers", "Adrian Jeremy", "dnp3"));
        configList.add(DataGenerator.buildJSONRequestStr("service_option", "anonymousftp", "on", ""));
        request = DataGenerator.buildFullJSONRequestStr("LINUX", "linuxa", configList);
        assertTrue(operateAndCompare(request, true));
    }    
}