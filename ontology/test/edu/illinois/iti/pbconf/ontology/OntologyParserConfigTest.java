/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;
import java.nio.file.Files;
import java.util.List;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.AfterClass;
import static org.junit.Assert.assertTrue;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import static org.junit.Assert.fail;
import org.semanticweb.owlapi.model.OWLException;

/**
 *
 * @author developer
 */
public class OntologyParserConfigTest {
    static Logger logger = Logger.getLogger(OntologyParserConfigTest.class.getName().replaceFirst(".+\\.",""));
    
    private static boolean setUpIsDone = false;
    
    /**
     * Setup the ontology for the parser test
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
     * Run setup one time through to load required configuration data for testing
     * @throws Exception
     */
    @Before
    public void setUp() throws Exception {
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
     * Execute tests of the ontology parser using both configuration and policy
     */
    public OntologyParserConfigTest() {}
    
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
     * Read in a test json file which contains an array of operations
     * Process operations and return state
     * 
     * @param filename
     * @return 
     */
    private Boolean loadAndTestJSONFile(String filename) {
        Boolean fileValid = true;
        StringBuilder sb = new StringBuilder();
        byte[] buffer = new byte[1024];
        InputStream istream = OntologyParserConfigTest.class.getResourceAsStream(filename);
        if (istream == null) {
            logger.error("Resource not found"+filename);
            return false;
        }
        try {
            while(istream.read(buffer) != -1) {
                sb.append(new String(buffer,"UTF-8"));
            }
            String jsonStr = sb.toString();
            
            JSONArray jArr = new JSONArray(jsonStr);
            for (int i = 0; i < jArr.length(); i++) {
                JSONObject jObj = jArr.getJSONObject(i);
                String expectedResult = jObj.getString("result");
                String req = jObj.getJSONObject("input").toString();
                OntologyParser parser = new OntologyParser();
                JSONObject result = parser.processTestInput(req);
                String outputStr = result.getString("output");
                
                JSONObject output = null;
                String status = "";               
                        
                try {
                    output = new JSONObject(outputStr);
                    status = output.getString("status");
                } catch (JSONException ex) {
                    logger.info(ex.toString());
                }
                
                fileValid = (fileValid && (expectedResult.equals(status)));         
            }
            
            return fileValid;
        } catch (IOException ex) {            
            logger.error("io",ex);
        }
        finally {
            try {
                istream.close();
            } catch (IOException exc) {
                logger.error("io",exc);
            }
        }
        return false;
    }

    /**
     * When you need more fine-grained testing. 
     * Loads a single JSON request from a file (as it comes from PBCONF).
     * Process request and return resulting "status".
     * 
     * @param filename
     * @return status ie. "VALID", "INVALID"
     */
    private String loadAndTestOneJSONRequest(String filename) throws Exception {
        URL fileURL = OntologyParserConfigTest.class.getResource(filename);
        
        File jsonFile = new File(fileURL.toURI());
        byte[] encoded = Files.readAllBytes(jsonFile.toPath());
        String req = new String(encoded, "UTF-8");
        
        OntologyParser parser = new OntologyParser();
        JSONObject result = parser.processInput(req);
        String outputStr = result.getString("output");
        JSONObject output = new JSONObject(outputStr);
        String status = output.getString("status");
        return status;
    }
    
    /**
     *
     * @throws Exception
     */
    @Test
    public void testTestResource() throws Exception {
        URL fileURL = OntologyParserConfigTest.class.getResource("configTest/configA.json");
        assertTrue(fileURL.toString().endsWith("edu/illinois/iti/pbconf/ontology/configTest/configA.json"));
        
        fileURL = OntologyParserConfigTest.class.getResource("configTest/NOTHERE.json");
        assertTrue(fileURL==null);
    }
    
    /**
     * Test the ability for OntologyParser to use backup/restore to restore 
     * consistency after a failure.
     * 
     * @throws Exception
     */
    @Test
    public void testOntologyRecover() throws Exception {
        String status = loadAndTestOneJSONRequest("configTest/oneBadConfig.json");
        assertTrue(status.equals("INVALID"));
        // Did it recover? 
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Test access configuration
     * @throws Exception
     */
    @Test
    public void testConfigAccess() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configAccess.json"));
    }
    
    /**
     * Test order dependant creation of IP address object using config byte set
     * It should set all bytes, then set type
     * followed by setting relation of object properties with sel421FAKEDEVICEB
     * @throws Exception 
     */
    @Test
    public void testConfigIPAddr() throws Exception { 
        assertTrue(loadAndTestJSONFile("configTest/configIPAddr_a.json"));
        assertTrue(loadAndTestJSONFile("configTest/configIPAddr_b.json"));
    }
    
    /**
     * Test set of mac address
     * @throws Exception
     */
    @Test
    public void testConfigMAC() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configMAC.json"));
    }
    
    /**
     * Test configuration of router
     * @throws Exception
     */
    @Test
    public void testConfigRouter() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configRouter.json"));
    }
    
    /**
     * Test hasFTPStt and hasFTPAnonStt
     * @throws Exception
     */
    @Test
    public void testConfigFTPAndFTPAnon() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configFTP.json"));
    }
    
    /**
     * Test retry delay
     * @throws Exception
     */
    @Test
    public void testConfigRetry() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configRetry.json"));
    }

    /**
     * This tests a single object property
     * The hasTelnetStt value (object) should be an IRI to the individual,
     *  http://iti.illinois.edu/iti/pbconf/core#off
    
     * @throws Exception 
     */
    @Test
    public void testOneObjectPropertyOff() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);
        URL fileURL = OntologyParserConfigTest.class.getResource("configTest/oneObjectPropertyOff.json");
        
        File jsonFile = new File(fileURL.toURI());
        byte[] encoded = Files.readAllBytes(jsonFile.toPath());
        String req = new String(encoded, "UTF-8");
        String status = null;
        OntologyParser parser = new OntologyParser();
        try {
            
            JSONObject result = parser.processInput(req);
            String outputStr = result.getString("output");
            JSONObject output = new JSONObject(outputStr);
            status = output.getString("status");        
        } catch (Throwable t) {
            logger.error("PUKED AGAIN",t);
            fail("Unexplained error.");
        }
        assertTrue("VALID".equals(status));
        // Did it recover? 
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Test configuration of hasTelnetStt
     * @throws Exception
     */
    @Test
    public void testConfigTelnet() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configTelnet.json"));
    }
    
    /**
     * Test configuration of hasDNP3Stt
     * @throws Exception
     */
    @Test
    public void testConfigDNP() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configDNP.json"));
    }
    
    /**
     * Test configuration of hasIEC61850
     * @throws Exception
     */
    @Test
    public void testConfigIEC() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configIEC.json"));
    }
    
    /**
     * Test configuration of hasNTPStt
     * @throws Exception
     */
    @Test
    public void testConfigNTP() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configNTP.json"));
    }
    
    /**
     * Test configuration of hasPingStt
     * @throws Exception
     */
    @Test
    public void testConfigPing() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configPing.json"));
    }
    
    /**
     * Test full SEL421Port5 setup
     * @throws Exception
     */
    @Test
    public void testSEL421Port5Config() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configSEL421AllowedOperationCreate_a.json"));
        assertTrue(loadAndTestJSONFile("configTest/configSEL421AllowedOperationCreate_b.json"));
        assertTrue(loadAndTestJSONFile("configTest/configSEL421AllowedOperationCreate_c.json"));
        assertTrue(loadAndTestJSONFile("configTest/configSEL421Port5IPAddrCreate.json"));
        assertTrue(loadAndTestJSONFile("configTest/configSEL421Port5RouterIPAddrCreate.json"));
        
        assertTrue(loadAndTestJSONFile("configTest/configSEL421Port5Create.json"));
        assertTrue(loadAndTestJSONFile("configTest/configSEL421SEL421Create.json"));
    }
    
    /**
     * Test SEL421Port5 setup in one file
     * @throws Exception
     */
    @Test
    public void testSEL421dPort5Config() throws Exception {
        assertTrue(loadAndTestJSONFile("configTest/configSEL421sel421d.json"));
    }
}
