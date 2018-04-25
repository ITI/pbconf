/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import edu.illinois.iti.pbconf.ontology.OntologyConfig.JSONConfig;
import java.util.Map;
import org.apache.log4j.Logger;
import static org.junit.Assert.assertTrue;
import org.junit.Test;

/**
 *
 * @author developer
 */
public class OntologyConfigTest {
    static Logger logger = Logger.getLogger(OntologyConfigTest.class.getName().replaceFirst(".+\\.",""));
    
    /**
     * Verify that pbconf.json loads newly named key value pairs
     * @throws Exception 
     */
    @Test
    public void test_JSONConfig() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        assertTrue("owl".equals(config.get("ontologyDirectory")));
        assertTrue("pbconf".equals(config.get("coreOntology")));       
        assertTrue(config.getAdditionalOntologies().size() > 0);
        assertTrue(config.getPrefixes().size() > 0);
    }
    
    /**
     * Verify that using an empty config file loads the base configuration
     * @throws Exception 
     */
    @Test
    public void test_JSONConfig_empty() throws Exception {
        OntologyConfig.JSONConfig config = new JSONConfig();
        assertTrue("owl".equals(config.get("ontologyDirectory")));
        assertTrue("pbconf".equals(config.get("coreOntology")));
        assertTrue(config.getAdditionalOntologies().isEmpty());
        assertTrue(config.getPrefixes().size() > 0);
    }
    
    /**
     * Verify that using an non-existent file falls back to the base configuration
     * @throws Exception 
     */
    @Test
    public void test_JSONConfig_no_exist() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("INVALID--FILE--NAME.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/INVALID__FILE__NAME.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        assertTrue("".equals(config.get("ontologyDirectory")));
        assertTrue("".equals(config.get("coreOntology")));
        assertTrue(config.getAdditionalOntologies().isEmpty());
        assertTrue(config.getPrefixes().isEmpty());
    }
    
    /**
     * Verify that using an invalid JSON file falls back to the base configuration
     * @throws Exception 
     */
    @Test
    public void test_JSONConfig_invalid() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("invalidJSON.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/invalidJSON.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        assertTrue("".equals(config.get("ontologyDirectory")));
        assertTrue("".equals(config.get("coreOntology")));
        assertTrue(config.getAdditionalOntologies().isEmpty());
        assertTrue(config.getPrefixes().isEmpty());
    }
    
    /**
     * Verify that prefixes and additional ontologies are correct
     * @throws Exception 
     */
    @Test
    public void test_ConfigPrefixesAdditional() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        assertTrue("owl".equals(config.get("ontologyDirectory")));
        assertTrue("pbconf".equals(config.get("coreOntology")));
        assertTrue(config.getAdditionalOntologies().size() > 0);
        assertTrue(config.getPrefixes().size() > 0);
        
        for (Map.Entry<String, String> prefix : config.getPrefixes().entrySet()) {
            if ("config".equals(prefix.getKey())) {
                assertTrue("http://iti.illinois.edu/iti/pbconf/config".equals(prefix.getValue()));
            }
            if ("policy".equals(prefix.getKey())) {
                assertTrue("http://iti.illinois.edu/iti/pbconf/policy".equals(prefix.getValue()));
            }
        }
    }
    
    /**
     * Make sure we can create / update additional properties in the configuration
     * @throws Exception 
     */
    @Test
    public void test_addUpdateProperty() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        //Set a value and confirm
        config.set("foo", "bar");
        assertTrue("bar".equals(config.get("foo")));
        
        //Update that value and reconfirm
        config.set("foo", "notbar");
        assertTrue("notbar".equals(config.get("foo")));
    }
    
    /**
     * Make sure we can set and remove a non-essential property
     * Make sure we can't remove an essential property
     * @throws Exception 
     */
    @Test
    public void test_removeProperty() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        //Set a value and confirm
        config.set("foo", "bar");
        assertTrue("bar".equals(config.get("foo")));
        
        boolean removeTest = config.remove("foo");
        assertTrue(removeTest == true);
        
        boolean failedRemove = config.remove("foo");
        assertTrue(failedRemove == false);   
        
        boolean skippedRemove = config.remove("coreOntology");
        assertTrue(skippedRemove == false);
    }
    
    /**
     * Get a map of reasoners available to us.
     * @throws Exception
     */
    @Test
    public void test_getClosedWorldReasoners() throws Exception {
//        OntologyConfig.JSONConfig config = new JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig config = new OntologyConfig.JSONConfig(c);
        
        Map<String, String> reasoners = config.getClosedWorldReasoners();
        assertTrue(reasoners.get("IndividualMustExist") != null);
    }
}
