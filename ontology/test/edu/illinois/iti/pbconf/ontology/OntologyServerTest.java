/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.io.UnsupportedEncodingException;
import java.lang.Thread.State;
import java.net.ServerSocket;
import java.net.Socket;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.junit.After;
import static org.junit.Assert.assertTrue;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;

/**
 *
 * @author andreobrown
 */
public class OntologyServerTest {
    static Logger logger = Logger.getLogger(OntologyServerTest.class.getName().replaceFirst(".+\\.",""));
    
    /**
     * Setup basic configurator
     */
    @BeforeClass 
    public static void once() {
        // setup logger
        BasicConfigurator.configure();
    }

    /**
     * Setup output streams to redirect console output from server to a byte
     * array. Byte array is read in tests to compare that the output is as
     * expected.
     */
    private final ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
    private final ByteArrayOutputStream errorStream = new ByteArrayOutputStream();

    /**
     * Default constructor
     */
    public OntologyServerTest() {
    }

    /**
     * Create output and error streams
     */
    @Before
    public void setUpStreams() {
        try {
            System.setOut(new PrintStream(outputStream,true,"UTF-8"));
            System.setErr(new PrintStream(errorStream,true,"UTF-8"));
        } catch (UnsupportedEncodingException ex) {
            logger.error(ex);
        }
    }

    /**
     * Cleanup streams on exit
     */
    @After
    public void tearDownStreams() {
        System.setOut(null);
        System.setErr(null);
    }

    /**
     * Test of main method, of class PBCONFServer.
     * @throws java.lang.Exception
     */
    @Test
    public void testServerStartup() throws Exception {
        //String[] args = null;
        int clientNumber = 0;
        // the ServerSocket used by the OntologyServer
        ServerSocket listener = new ServerSocket(9090);
        // the client socket connecting to that server
        Socket socket = new Socket("localhost", 9090);
        assertTrue(socket.isConnected());
        
        // new OntologyServer will spin up with client connects, (and listener.accept() returns)
        OntologyServer server = new OntologyServer(listener.accept(), clientNumber++);
        State st = server.getState();
        assertTrue(st != null);
        assertTrue(clientNumber == 1); // these silly asserts keep findbugs happy
   
        String serverOut = outputStream.toString("UTF-8");
        logger.info("Server output: "+serverOut);
        assertTrue(outputStream.toString("UTF-8").contains("OntologyServer  - New connection with client #0"));
    }
}
