/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.junit.Test;

/**
 *
 * @author andreobrown
 */
public class OntologyClientTest {
    static Logger logger = Logger.getLogger(OntologyClientTest.class.getName().replaceFirst(".+\\.",""));  
    
    /**
     * All we should need to set up is a basic configuration for the client tests
     */
    public OntologyClientTest() {
        // setup logger
        BasicConfigurator.configure();
    }

    /**
     * Test of main method, of class OntologyClient.
     * @throws java.lang.Exception
     */
    @Test
    public void testMain() throws Exception {
        logger.info("main");
        //String[] args = null;
        
        logger.info("OntologyClient test disabled.");
        //OntologyClient.main(args);
        // TODO review the generated test code and remove the default call to fail.
    }

}
