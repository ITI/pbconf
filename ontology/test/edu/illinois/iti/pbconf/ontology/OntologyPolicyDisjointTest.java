/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import java.io.FileNotFoundException;
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
import org.semanticweb.owlapi.model.OWLAxiom;

/**
 * Test the disjointWith statement for string handling
 * @author Joe DiGiovanna
 */
public class OntologyPolicyDisjointTest {
    static Logger logger = Logger.getLogger(OntologyPolicyDisjointTest.class.getName().replaceFirst(".+\\.",""));
    
    private static boolean setUpIsDone = false;
    
    /**
     * Setup configurator
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
     * Quick save by prefix
     * @param prefix 
     */
    public void save(String prefix) {
        try {
            Ontology.instance().saveOntologyByPrefix(prefix);
        } catch (OWLException | FileNotFoundException ex) {
            logger.info(ex.toString());
        }
    }
    
    /**
     * Make sure the character converter is working
     * @throws Exception 
     */
    @Test
    public void char_convert_test() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        
        String aa = ope.convertCount(0);
        String ZZ = ope.convertCount(2703);
        
        assertTrue("aa".equals(aa));
        assertTrue("ZZ".equals(ZZ));       
    }

    /**
     * Test setting so that SEL421 class cannot have both Joe and Andy authorized for DNP3 access
     * @throws Exception 
     */
    @Test
    public void sel421_disjoint_test_0() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("SEL421", "disjointWith", "(authDNP $value Andy) $and (authDNP $value Joe)");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Test setting so that SEL421 class cannot have either Joe or Andy authorized for DNP3 access
     * @throws Exception 
     */
    @Test
    public void sel421_disjoint_test_1() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("SEL421", "disjointWith", "$not ((authDNP $value Andy) $or (authDNP $value Joe))");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * SEL421 class cannot authorize DNP3 service to something of class IPAddress
     * @throws Exception 
     */
    @Test
    public void sel421_disjoint_test_2() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("SEL421", "disjointWith", "authDNP $some IPAddress");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * SEL421 class cannot authorize DNP3 service to something of class IPAddress
     * @throws Exception 
     */
    @Test
    public void sel421_disjoint_test_3() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("LINUX", "disjointWith", "authDNP $some Person");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(!"".equals(Ontology.instance().validate("test")));
    }
}