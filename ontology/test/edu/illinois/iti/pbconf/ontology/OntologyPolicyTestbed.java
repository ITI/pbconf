/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import java.io.FileNotFoundException;
import java.util.List;
import java.util.logging.Level;
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

/**
 * Test the equivalentTo statement for string handling
 * @author Joe DiGiovanna
 */
public class OntologyPolicyTestbed {
    static Logger logger = Logger.getLogger(OntologyPolicyTestbed.class.getName().replaceFirst(".+\\.",""));
    
    private boolean setUpIsDone = false;
    
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
        if (setUpIsDone == false) {
            setUpIsDone = true;
        
            Ontology.instance().reset();
//            OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
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

            /**
             * This is so that we can add policy on existing configuration to test it.
             */
            
            /*
            List dataArray = new ArrayList();

            dataArray.add(buildJSONRequestStr("LINUX", "linuxa", "type", "", "LINUX", ""));
            dataArray.add(buildJSONRequestStr("LINUX", "Joe", "type", "", "Person", ""));
            dataArray.add(buildJSONRequestStr("LINUX", "Adrian", "type", "", "Person", ""));
            dataArray.add(buildJSONRequestStr("LINUX", "Andy", "type", "", "Person", ""));
            dataArray.add(buildJSONRequestStr("LINUX", "Jeremy", "type", "", "Person", ""));
            dataArray.add(buildJSONRequestStr("LINUX", "linuxa", "service", "authusers", "on", ""));
            dataArray.add(buildJSONRequestStr("LINUX", "linuxa", "service", "anonymousftp", "on", ""));
            dataArray.add(buildJSONRequestStr("SEL421", "sel421FAKEDEVICEB", "service_option", "authusers", "Joe Andy Adrian Jeremy", "telnet"));
            dataArray.add(buildJSONRequestStr("LINUX", "linuxa", "service_option", "authusers", "Joe Andy Adrian Jeremy", "ftp"));
            dataArray.add(buildJSONRequestStr("LINUX", "linuxa", "service_option", "authusers", "Adrian Jeremy", "dnp3"));
            dataArray.add(buildJSONRequestStr("LINUX", "linuxa", "service_option", "anonymousftp", "on", ""));

            String dataStatement = buildFullJSONRequestStr(dataArray);

            assertTrue(operateAndCompare(dataStatement, true));
            */
        }
    }
    
    /**
     * Quick save by prefix
     * @param prefix 
     */
    public void save(String prefix) {
        try {
            Ontology.instance().saveOntologyByPrefix(prefix);
        } catch (OWLException | FileNotFoundException ex) {
            java.util.logging.Logger.getLogger(OntologyPolicyStatementTest.class.getName()).log(Level.SEVERE, null, ex);
        }
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
     * Build out a full policy statement to test against configuration data
     * @param ontologizer
     * @param dataArray
     * @return 
     */
    public String buildFullJSONPolicyRequestStr(String ontologizer, List dataArray) {       
        String jsonStr = "{\"ontology\":\"policy\",\"data\":[[";
        
        for (int i = 0; i < dataArray.size(); i++) {
            jsonStr = jsonStr.concat((String) dataArray.get(i));
            if (i != dataArray.size() - 1) {
                jsonStr = jsonStr.concat(",");
            }
        }
        
        jsonStr += "]]}";
        
        return jsonStr;
    }
    
    /**
     * Get a single data item
     * @param cls
     * @param axioms
     * @return 
     */
    public String buildJSONPolicyRequestStr(String cls, String[] axioms) {
        String jsonStr = "{\"Class\":\"" + cls + "\",\"Axioms\":[";
        for (int i = 0; i < axioms.length; i++) {
            if (i != 0) {
                jsonStr += ",";
            }
            jsonStr = jsonStr + axioms[i];
        }
        jsonStr += "]}";
        return jsonStr;
    }
    
    /**
     * Use the processInput function in OntologyParser to process the JSON data
     * This can accept both config and policy statements
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
     * Set a valid disjoint statement that implies we're disjoint with something calle FAKECLASS
     * @throws Exception 
     */
    @Test
    public void sel421_statement_test_0() throws Exception {
       /**
        * This is purely for easy testing of axiom layout
        */
        
        //Quick test to make sure replace ontology is working
        Ontology.instance().replaceOntology("policy");
        Ontology.instance().replaceOntology("config");
        
        assertTrue(Ontology.instance().isConstistent());
               
        /*
        List dataArray = new ArrayList();

        String[] classAxioms = new String[20];
        String[] individualAxioms = new String[20];
        String[] fakeIndividualAxioms = new String[20];
        String[] objectPropertyAxioms = new String[20];
        String[] dataPropertyAxioms = new String[20];
        
        String testTargetClass = "SEL421";
        String testTargetIndividual = "sel421FAKEDEVICEA";
        String testTargetFakeIndividual = "sel421fake";
        String testTargetObjectProperty = "hasNTPStt";
        String testTargetDataProperty = "hasLvl2Pwd";
        
        classAxioms[0] = "{\"s\":\"SEL421\",\"p\":\"mustNotHave\",\"o\":\"hasPingStt\"}";
        classAxioms[1] = "{\"s\":\"SEL421\",\"p\":\"requires\",\"o\":\"password.level2\"}";
      
        classAxioms[2] = "{\"s\":\"password.level2\",\"p\":\"min-length\",\"o\":\"4\"}";
        classAxioms[3] = "{\"s\":\"password.level2\",\"p\":\"max-length\",\"o\":\"12\"}";
        classAxioms[4] = "{\"s\":\"password.level2\",\"p\":\"complexity\",\"o\":\"MIXEDCASE\"}";

        classAxioms[5] = "{\"s\":\"password.level1A\",\"p\":\"min-length\",\"o\":\"6\"}";
        classAxioms[6] = "{\"s\":\"password.level1A\",\"p\":\"max-length\",\"o\":\"12\"}";
        classAxioms[7] = "{\"s\":\"password.level1A\",\"p\":\"complexity\",\"o\":\"LOWERCASE\"}";
        
        classAxioms[8] = "{\"s\":\"password.level1B\",\"p\":\"min-length\",\"o\":\"8\"}";
        classAxioms[9] = "{\"s\":\"password.level1B\",\"p\":\"max-length\",\"o\":\"16\"}";
        classAxioms[10] = "{\"s\":\"password.level1B\",\"p\":\"complexity\",\"o\":\"UPPERCASE\"}";

        individualAxioms[0] = "{\"s\":\"sel421FAKEDEVICEA\",\"p\":\"isA\",\"o\":\"SEL421\"}";
        individualAxioms[1] = "{\"s\":\"sel421\",\"p\":\"mustNotHave\",\"o\":\"password.level2\"}";       
        individualAxioms[2] = "{\"s\":\"hasNTPStt\",\"p\":\"state\",\"o\":\"on\"}";
        individualAxioms[3] = "{\"s\":\"hasDNP3Stt\",\"p\":\"status\",\"o\":\"off\"}";
        
        fakeIndividualAxioms[0] = "{\"s\":\"sel421fake\",\"p\":\"mustNotHave\",\"o\":\"password.level2\"}"; 
        
        dataPropertyAxioms[0] = "{\"s\":\"password.level1A\",\"p\":\"min-length\",\"o\":\"6\"}";
        dataPropertyAxioms[1] = "{\"s\":\"password.level1A\",\"p\":\"max-length\",\"o\":\"12\"}";
        dataPropertyAxioms[2] = "{\"s\":\"password.level1A\",\"p\":\"complexity\",\"o\":\"LOWERCASE\"}";
        
        dataPropertyAxioms[3] = "{\"s\":\"password.level1B\",\"p\":\"min-length\",\"o\":\"8\"}";
        dataPropertyAxioms[4] = "{\"s\":\"password.level1B\",\"p\":\"max-length\",\"o\":\"16\"}";
        dataPropertyAxioms[5] = "{\"s\":\"password.level1B\",\"p\":\"complexity\",\"o\":\"UPPERCASE\"}";
        
        
        PBConfParser parser = new PBConfParser();
        
        for (String classAxiom : classAxioms) {
            JSONObject obj = new JSONObject(classAxiom);
            obj.put("s", PBConfParser.translateProperty(obj.getString("s")));
            obj.put("o", PBConfParser.translateProperty(obj.getString("o")));
            logger.info("axiom : " + obj.toString());
            logger.info("target = " + testTargetClass + ", subject = " + obj.get("s"));
            logger.info("predicate = " + obj.getString("p") + ", object = " + obj.getString("o"));
            logger.info("target = " + parser.classifyTarget(testTargetClass, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("subject = " + parser.classifySubject(testTargetClass, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("object = " + parser.classifyObject(testTargetClass, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("policytype = " + parser.getPolicyType(testTargetClass, obj.getString("s"), obj.getString("p"), obj.getString("o")));       
        }
        
        for (String individualAxiom : individualAxioms) {
            JSONObject obj = new JSONObject(individualAxiom);
            obj.put("s", PBConfParser.translateProperty(obj.getString("s")));
            obj.put("o", PBConfParser.translateProperty(obj.getString("o")));
            logger.info("axiom : " + obj.toString());
            logger.info("target = " + testTargetIndividual + ", subject = " + obj.get("s"));
            logger.info("predicate = " + obj.getString("p") + ", object = " + obj.getString("o"));
            logger.info("target = " + parser.classifyTarget(testTargetIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("subject = " + parser.classifySubject(testTargetIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("object = " + parser.classifyObject(testTargetIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("policytype = " + parser.getPolicyType(testTargetIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));       
        }
        
        for (String individualAxiom : fakeIndividualAxioms) {
            JSONObject obj = new JSONObject(individualAxiom);
            obj.put("s", PBConfParser.translateProperty(obj.getString("s")));
            obj.put("o", PBConfParser.translateProperty(obj.getString("o")));
            logger.info("axiom : " + obj.toString());
            logger.info("target = " + testTargetIndividual + ", subject = " + obj.get("s"));
            logger.info("predicate = " + obj.getString("p") + ", object = " + obj.getString("o"));
            logger.info("target = " + parser.classifyTarget(testTargetFakeIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("subject = " + parser.classifySubject(testTargetFakeIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("object = " + parser.classifyObject(testTargetFakeIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("policytype = " + parser.getPolicyType(testTargetFakeIndividual, obj.getString("s"), obj.getString("p"), obj.getString("o")));       
        }
        
        for (String objectPropertyAxiom : objectPropertyAxioms) {
            JSONObject obj = new JSONObject(objectPropertyAxiom);
            obj.put("s", PBConfParser.translateProperty(obj.getString("s")));
            obj.put("o", PBConfParser.translateProperty(obj.getString("o")));
            logger.info("axiom : " + obj.toString());
            logger.info("target = " + testTargetIndividual + ", subject = " + obj.get("s"));
            logger.info("predicate = " + obj.getString("p") + ", object = " + obj.getString("o"));
            logger.info("target = " + parser.classifyTarget(testTargetObjectProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("subject = " + parser.classifySubject(testTargetObjectProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("object = " + parser.classifyObject(testTargetObjectProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("policytype = " + parser.getPolicyType(testTargetObjectProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));       
        }
             
        for (String dataPropertyAxiom : dataPropertyAxioms) {
            JSONObject obj = new JSONObject(dataPropertyAxiom);
            obj.put("s", PBConfParser.translateProperty(obj.getString("s")));
            obj.put("o", PBConfParser.translateProperty(obj.getString("o")));
            logger.info("axiom : " + obj.toString());
            logger.info("target = " + testTargetIndividual + ", subject = " + obj.get("s"));
            logger.info("predicate = " + obj.getString("p") + ", object = " + obj.getString("o"));
            logger.info("target = " + parser.classifyTarget(testTargetDataProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("subject = " + parser.classifySubject(testTargetDataProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("object = " + parser.classifyObject(testTargetDataProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));
            logger.info("policytype = " + parser.getPolicyType(testTargetDataProperty, obj.getString("s"), obj.getString("p"), obj.getString("o")));       
        }
        */
        
    }
    
   
}