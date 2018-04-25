package edu.illinois.iti.pbconf.ontology;

import java.io.BufferedReader;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.Date;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.OWLException;

/**
 * PBCONF server listens on port 9090 (see PBCONFClient)
 */
public class OntologyServer extends Thread {
    private Socket socket;
    private int clientNumber;
    private static final int DEFAULT_PORT = 9090;
    
    static Logger logger = Logger.getLogger(OntologyServer.class.getName().replaceFirst(".+\\.",""));
    static {
        BasicConfigurator.configure();
    }
    
    /**
     * Entry point for ontology server
     * @param args
     * @throws Exception
     */
    public static void main(String[] args) throws Exception {
        int clientNumber = 0;
        try (ServerSocket listener = new ServerSocket(DEFAULT_PORT)) {
            while (true) {
                new OntologyServer(listener.accept(), clientNumber++).start();
            }
        }
    }
    
    /**
     * Each client now gets their own ontology instance.  
     * This should allow for better policy handling and no overlap.
     * TODO: It is also enormously inefficient an inelegant. We should prioritize 
     * a better solution.
     * @param socket
     * @param clientNumber 
     */
    public OntologyServer(Socket socket, int clientNumber) {
        this.socket = socket;
        this.clientNumber = clientNumber;
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("/etc/pbconf/ontology/pbconf.json");      
        logger.info("New connection with client #" + clientNumber + " at " + socket);
        try {
            Ontology.instance().initialize(cfg);
        } catch (OWLException | FileNotFoundException ex) {
            logger.error("Ontology initialization error: ",ex);
        }
    }
    
    @Override
    public void run() {
        try {
            BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream(),"UTF-8"));
            OutputStreamWriter osw = new OutputStreamWriter(socket.getOutputStream(),"UTF-8");
            PrintWriter out = new PrintWriter(osw, true);
            OntologyParser ontologyParser = new OntologyParser();
            out.println("Connection established"); // welcome message

            while (true) {
                String input, output = "";
                input = in.readLine();
                logger.info("[" + new Date().toString() + "] Recv: " + input);
                
                JSONObject resultObj = ontologyParser.processInput(input);
                output = resultObj.getString("output");
                Boolean exiting = resultObj.getBoolean("exiting");
                
                if (exiting == true) {
                    break;
                } else {
                    out.println(output);
                }
            }
        } catch (IOException e) {
            logger.info("Error handling client# " + clientNumber + ": " + e);
        } finally {
            try {
                socket.close();
            } catch (IOException e) {
                logger.info("Couldn't close a socket, what's going on?");
            }
            logger.info("Connection with client# " + clientNumber + " closed");
        }
    }
}
