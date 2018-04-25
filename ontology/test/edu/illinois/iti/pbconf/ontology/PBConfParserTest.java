package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import edu.illinois.iti.pbconf.ontology.OntologyConfig.JSONConfig;
import java.io.FileNotFoundException;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.junit.AfterClass;
import static org.junit.Assert.assertTrue;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.semanticweb.owlapi.model.OWLException;

/**
 *
 * @author developer
 */
public class PBConfParserTest {
    static Logger logger = Logger.getLogger(PBConfParserTest.class.getName().replaceFirst(".+\\.",""));
    
    /**
     * Setup basic configurator
     */
    @BeforeClass 
    public static void once() {
        // setup logger
        BasicConfigurator.configure();
        logger.info("Configured, starting PBConfParserTest");
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
     * Before each test, reinitialize
     * @throws Exception
     */
    @Before
    public void setUp() throws Exception {
        Ontology.instance().reset();
//        JSONConfig cfg = new JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        JSONConfig cfg = new JSONConfig(c);
        
        Ontology.instance().initialize(cfg);
        logger.info("Initialized with full pbconf / config / policy ontologies");
    }
    
    /**
     * Test parsing of pbconf property password.level1, mixedcase
     * @throws Exception 
     */
    @Test
    public void testLevel1Pwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1", "complexity", "MIXEDCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1Pwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of pbconf property password.level2, lowercase
     * @throws Exception 
     */
    @Test
    public void testLevel2Pwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level2", "complexity", "LOWERCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl2Pwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of pbconf property password.levelC, uppercase
     * @throws Exception 
     */
    @Test
    public void testLevelCPwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.levelC", "complexity", "UPPERCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvlCPwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of pbconf property password.level1A
     * @throws Exception 
     */
    @Test
    public void testLevel1APwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1A", "complexity", "MIXEDCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1APwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of pbconf property password.level1B
     * @throws Exception 
     */
    @Test
    public void testLevel1BPwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1B", "complexity", "MIXEDCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1BPwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of pbconf property password.level1O
     * @throws Exception 
     */
    @Test
    public void testLevel1OPwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1O", "complexity", "MIXEDCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1OPwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of pbconf property password.level1P
     * @throws Exception 
     */
    @Test
    public void testLevel1PPwd() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1P", "complexity", "MIXEDCASE");
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1PPwd".equals(json.getString("translatedSubject")));
    }
    
    /**
     * Test parsing of value property
     * @throws Exception 
     */
    @Test
    public void testAxiomSelA() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1P", "complexity", "MIXEDCASE");    
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1PPwd".equals(json.getString("translatedSubject")));
        assertTrue("MIXEDCASE".equals(json.getString("value")));
    }
    
    /**
     * Test parsing of value property
     * @throws Exception 
     */
    @Test
    public void testAxiomSelB() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1P", "complexity", "LOWERCASE");    
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1PPwd".equals(json.getString("translatedSubject")));
        assertTrue("LOWERCASE".equals(json.getString("value")));
    }
    
    /**
     * Test parsing of value property
     * @throws Exception 
     */
    @Test
    public void testAxiomSelC() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("SEL421", "password.level1P", "complexity", "UPPERCASE");    
        assertTrue("cwr".equals(json.getString("policyType")));
        assertTrue("hasLvl1PPwd".equals(json.getString("translatedSubject")));
        assertTrue("UPPERCASE".equals(json.getString("value")));
    }
    
    /**
     * Test axiom data
     * @throws Exception
     */
    @Test
    public void testAxiomData() throws Exception {
        PBConfParser pbParser = new PBConfParser();
        JSONObject json = pbParser.getAxiomParts("password.level1P", "password.level1P", "complexity", "UPPERCASE");    
        assertTrue("data".equals(json.getString("policyType")));
        assertTrue("hasLvl1PPwd".equals(json.getString("translatedTarget")));
        assertTrue("hasLvl1PPwd".equals(json.getString("translatedSubject")));
        assertTrue("complexity".equals(json.getString("predicate")));
        assertTrue("UPPERCASE".equals(json.getString("value")));
    }
}
