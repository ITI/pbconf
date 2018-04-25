/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStream;
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
import org.semanticweb.owlapi.model.OWLException;

/**
 *
 * @author developer
 */
public class OntologyParserPolicyTest {

    static Logger logger = Logger.getLogger(OntologyParserPolicyTest.class.getName().replaceFirst(".+\\.", ""));

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
     * Load ontologies and test data
     *
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
            assertTrue(operateAndCompare((String) dataStr, true));
        }

        setUpIsDone = true;
    }

    /**
     * Use the processInput function in OntologyParser to process the JSON data
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
     * Execute tests of the ontology parser using both configuration and policy
     */
    public OntologyParserPolicyTest() {
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
            logger.info(ex.toString());
        }
    }

    /**
     * Read in a test json file which contains an array of operations Process
     * operations and return state
     *
     * @param filename
     * @return
     */
    private Boolean loadAndTestJSONFile(String filename) {
        Boolean fileValid = true;
        StringBuilder sb = new StringBuilder();
        byte[] buffer = new byte[1024];
        InputStream istream = OntologyParserPolicyTest.class.getResourceAsStream(filename);
        if (istream == null) {
            logger.error("Resource not found" + filename);
            return false;
        }
        try {
            while (istream.read(buffer) != -1) {
                sb.append(new String(buffer, "UTF-8"));
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
            logger.error("io", ex);
        } finally {
            try {
                istream.close();
            } catch (IOException exc) {
                logger.error("io", exc);
            }
        }
        return false;
    }

    /**
     * Set policy on sel421FAKEDEVICEA unit individual
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyIndividualPwd() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyIndividualPwd.json"));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on sel421FAKEDEVICEA unit individual
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyIndividualNotPwd() throws Exception {
        Ontology.instance().prepareForTesting();

        assertTrue(loadAndTestJSONFile("policyTest/policyIndividualNotPwd.json"));

        assertTrue(!"".equals(Ontology.instance().validate("test")));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on sel421FAKEDEVICEA unit individual
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyIndividualRequirements() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyIndividualRequirements.json"));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on SEL421 class
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyClassRequirements() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyClassRequirements.json"));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on SEL421 class
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyClassRequirementsInvalid() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyClassRequirementsInvalid.json"));
        
        assertTrue(!"".equals(Ontology.instance().validate("test")));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on SEL421 class
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyClassRequirementsInvalidB() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyClassRequirementsInvalidStatus.json"));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on sel421 units requiring them to have a level 2 password
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyRequiresLvl2Pwd() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyLvl2PwdRequired.json"));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Set policy on sel421 units requiring them to have a level 2 password
     *
     * @throws java.lang.Exception
     */
    @Test
    public void testPolicyRequiresLvlCPwd() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyLvlCPwdRequired.json"));

        assertTrue(ClosedWorld.clearDuplicateClosedWorldReasoners("TESTClosedWorldReasoner", "tcwr"));
    }

    /**
     * Should allow this because nothing uses level 1b passwords
     *
     * @throws Exception
     */
    @Test
    public void testPolicyDataPropLvl1BPwd() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyLvl1BPwd.json"));
    }

    /**
     * Should throw invalid because level 2 passwords exist that aren't
     * lowercase
     *
     * @throws Exception
     */
    @Test
    public void testPolicyDataPropLvl2PwdInvalid() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyLvl2Pwd.json"));
        
        assertTrue(!"".equals(Ontology.instance().validate("test")));
    }

    /**
     * Should be valid. Basically previous test but allow for mixed case
     *
     * @throws Exception
     */
    @Test
    public void testPolicyDataPropLvl2PwdValid() throws Exception {
        assertTrue(loadAndTestJSONFile("policyTest/policyLvl2PwdValid.json"));
    }
}
