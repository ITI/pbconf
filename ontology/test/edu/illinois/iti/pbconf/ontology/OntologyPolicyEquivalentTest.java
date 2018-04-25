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
 * Test the equivalentTo statement for string handling
 * @author Joe DiGiovanna
 */
public class OntologyPolicyEquivalentTest {
    static Logger logger = Logger.getLogger(OntologyPolicyEquivalentTest.class.getName().replaceFirst(".+\\.",""));
    
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
     * Build out a full configuration statement so we have real configuration data
     * @param dataArray
     * @return 
     */
    public String buildFullJSONRequestStr(List dataArray) {
        String jsonStr = "{\"ontology\":\"config\",\"properties\":[";
        
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
     * @param ontologizer
     * @param ind
     * @param Op
     * @param Key
     * @param Val
     * @param Svc
     * @return 
     */
    public String buildJSONRequestStr(String ontologizer, String ind, String Op, String Key, String Val, String Svc) {
        String jsonStr = "{\"ontologizer\":\"" + ontologizer + "\",\"individual\":\"" + ind + "\", \"properties\":["; 

        jsonStr += "{\"Op\":\"" + Op + "\",\"Key\":\"" + Key + "\",\"Val\":\"" + Val + "\", \"Svc\":\"" + Svc + "\"}";
        
        jsonStr += "]}";
        
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
     * Test out string conversion for encoding values
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
     * Make sure we can validate what is a status key
     * @throws Exception 
     */
    @Test
    public void ont_status_test() throws Exception {
        boolean testA = Ontologizer.getOntologizer("SEL421").isStatusKey("hasDNP3Stt");
        boolean testB = Ontologizer.getOntologizer("SEL421").isStatusKey("hasFTPStt");
        boolean testC = Ontologizer.getOntologizer("SEL421").isStatusKey("hasFTPAnonStt");
        boolean testD = Ontologizer.getOntologizer("SEL421").isStatusKey("hasRouterIPAddr");
        boolean testE = Ontologizer.getOntologizer("SEL421").isStatusKey("hasPort5");
        
        assertTrue(testA);
        assertTrue(testB);
        assertTrue(testC);
        assertFalse(testD);
        assertFalse(testE);
    }

    /**
     * Set some fake value demonstrating some of the parsing process
     * @throws Exception 
     */
    @Test
    public void sel421_equivalent_test_0() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("LINUX", "equivalentTo", "(authDNP $value Andy) $and (authDNP $value Joe)");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Set a valid requirement that authDNP is of type Person
     * @throws Exception 
     */
    @Test
    public void sel421_equivalent_test_1() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("LINUX", "equivalentTo", "authDNP $some Person");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Set a valid equivalency but then set an invalid negation.
     * @throws Exception 
     */
    @Test
    public void sel421_equivalent_test_2() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("LINUX", "equivalentTo", "(authDNP $some Person)");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(ope.isValid());
        assertTrue(Ontology.instance().isConstistent());
        
        //This part should fail because the previous statement isn't reconcilable with it.
        OWLAxiom axB = ope.constructAxiomFromString("LINUX", "equivalentTo", "$not (authDNP $some Person)");
        assertTrue(axB != null);
        ope.applyAxiom(axB);
        assertTrue(ope.isValid() == false);
        assertTrue("".equals(Ontology.instance().validate("test")));
    }
    
    /**
     * Set a valid equivalency but then set the same as disjoint.
     * @throws Exception 
     */
    @Test
    public void sel421_equivalent_test_3() throws Exception {
        OntologyPolicyStringEngine ope = new OntologyPolicyStringEngine();
        OWLAxiom ax = ope.constructAxiomFromString("LINUX", "equivalentTo", "(authDNP $some IPAddress)");
        assertTrue(ax != null);
        ope.applyAxiom(ax);
        assertTrue(ope.isValid());
        assertTrue(Ontology.instance().isConstistent());

        //This part should fail because the previous statement isn't reconcilable with it.
        OWLAxiom axB = ope.constructAxiomFromString("LINUX", "disjointWith", "(authDNP $some IPAddress)");
        assertTrue(axB != null);
        ope.applyAxiom(axB);
        assertTrue(ope.isValid());
        assertTrue(!"".equals(Ontology.instance().validate("test")));
    }
}