package edu.illinois.iti.pbconf.ontology;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.net.Socket;
import java.util.Date;

/**
 * Ontology client 
 * Manages connection between the Ontology server and PBConf
 * @author josephdigiovanna
 */
public class OntologyClient {

    /**
     * HTML new line character
     */
    public static final String NEW_LINE = "<br>";
    
    /**
     * Default host
     */
    public static final String DEFAULT_HOST = "localhost";
    
    /**
     * Default port
     */
    public static final int DEFAULT_PORT = 9090;

    /**
     * Main loop for managing input on port 9090
     * @param args
     * @throws Exception
     */
    public static void main(String[] args) throws Exception {
        Socket socket = new Socket(DEFAULT_HOST, DEFAULT_PORT);
        BufferedReader in = new BufferedReader(new InputStreamReader(socket.getInputStream(),"UTF-8"));
        BufferedReader console = new BufferedReader(new InputStreamReader(System.in,"UTF-8"));
        OutputStreamWriter osw = new OutputStreamWriter(socket.getOutputStream(),"UTF-8");
        PrintWriter out = new PrintWriter(osw, true);

        System.out.println(in.readLine()); // welcome message

        while (true) {
            String input, output;
            System.out.print(">>> ");
            output = console.readLine();
            out.println(output);
            System.out.println("[" + new Date().toString() + "] Send: " + output);
            input = in.readLine();
            if (input != null) {
                System.out.println(input.replace(NEW_LINE, "\n"));
            }
            
            while (in.ready()) {
                input = in.readLine();
                if (input != null) {
                    System.out.println(input.replace(NEW_LINE, "\n"));
                }
            }
            
            if (output != null && output.equals("exit")) {
                break;
            }
        }
    }
}
