/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import edu.illinois.iti.pbconf.ontology.validator.ClassMustHaveProperty;
import edu.illinois.iti.pbconf.ontology.validator.ClassMustNotHaveProperty;
import edu.illinois.iti.pbconf.ontology.validator.IndividualMustExist;
import java.io.FileNotFoundException;
import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.util.List;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.AfterClass;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLException;

/**
 *
 * @author anderson
 */
public class ClosedWorldTest {

    static final Logger logger = Logger.getLogger(ClosedWorldTest.class.getName().replaceFirst(".+\\.", ""));
   
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
     * Initialize test data on startup
     * @throws Exception
     */
    @Before
    public void setUp() throws Exception {
        if (setUpIsDone) {
            return;
        }
        
        Ontology.instance().reset();
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
        List data = DataGenerator.getTestConfigurationDataRequestCWR();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        
        setUpIsDone = true;
    }
    
    /**
     * After a test class, we will reset policy and configuration
     */
    @AfterClass
    public static void after() {
        //Clean everything up      
        Ontology.instance().reset();
        //OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
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
     * Newbie test to remind myself how to load classes by name.
     *
     * @throws Exception
     */
    @Test
    public void test_getClass() throws Exception {
        Ontology.instance().prepareForTesting();
        
        String className = "edu.illinois.iti.pbconf.ontology.validator.IndividualMustExist";
        ClassLoader loader = ClosedWorld.class.getClassLoader();
        Class imeClass = loader.loadClass(className);
        assertTrue(imeClass != null);
        assertTrue(imeClass.getName().equals(className));
    }
    
    /**
     * Load the IndividualMustExist class, then validate an individual known to exist.
     * @throws Exception 
     */
    @Test
    public void test_individualMustExist() throws  Exception {
        Ontology.instance().prepareForTesting();
        
        String className = "edu.illinois.iti.pbconf.ontology.validator.IndividualMustExist";
        ClassLoader loader = ClosedWorld.class.getClassLoader();
        Class imeClass = loader.loadClass(className);
        assertTrue(imeClass != null);
        assertTrue(imeClass.getName().equals(className));

        // The bootstrap config ontology contains an indiv "sel421FAKEDEVICEB", whose identity IRI is in the core namespace
        Individual test421a = Ontology.instance().getIndividual("test421a", Ontology.instance().getOntology("pbconf"));
        
        // get an instance of IME
        Constructor ctor;
        IndividualMustExist ime;
        boolean exists = false;
        try {
            ctor = imeClass.getConstructor();
            ime = (IndividualMustExist) ctor.newInstance();
            exists = ime.validateRequest(test421a);            
        } catch (NoSuchMethodException | SecurityException | IllegalArgumentException | InvocationTargetException ex) {
            fail(ex.toString());
        }
        assertTrue(exists);
    }
    
    /**
     * Get an instance of IndividualMustExist from ClosedWorld.
     * @throws Exception 
     */
    @Test
    public void test_getValidator() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("IndividualMustExist");
        assertTrue(validator instanceof IndividualMustExist);
    }
    
    /**
     * Validate a known individual against the "IndividualMustExist" registered validator.
     * @throws Exception 
     */
    @Test 
    public void test_validateIndividualMustExist() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("IndividualMustExist");
        assertTrue(validator instanceof IndividualMustExist);
        
        // The bootstrap config ontology contains an indiv "sel421FAKEDEVICEB", whose identity IRI is in the core namespace
        Individual testInd = Ontology.instance().getIndividual("test421a", Ontology.instance().getOntology("pbconf"));

        assertTrue(validator.validateRequest(testInd));        
    }
    
    /**
     * Validate known individuals against the "ClassMustHaveProperty" registered validator.
     * @throws Exception 
     */
    @Test 
    public void test_validateClassMustHaveProperty_a() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("ClassMustHaveProperty");
        assertTrue(validator instanceof ClassMustHaveProperty);
        
        validator.overrideClass("TESTClosedWorldReasoner", "tcwr");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI("TEST421"));
        OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1Pwd"));
 
        boolean isValid = validator.validateRequest(cls, odp);
        if (!isValid) {
            logger.info(validator.getExplanation());
        }
        assertTrue(isValid == true);
        
        if (isValid) {
            String nextName = ClosedWorld.getNextClosedWorldReasonerName(validator.getCurrentClass(), validator.getCurrentPrefix());
            boolean isSaved = validator.saveLastRequest(nextName);
            assertTrue(isSaved == true);

            boolean canReason = ClosedWorld.reason(validator.getCurrentClass(), validator.getCurrentPrefix());
            assertTrue(canReason == true);
        }       
    }
    
    /**
     * Validate known individuals against the "ClassMustHaveProperty" registered validator.
     * This one should fail with test421c not having a hasLvl2Pwd value set
     * @throws Exception 
     */
    @Test 
    public void test_validateClassMustHaveProperty_b() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("ClassMustHaveProperty");
        assertTrue(validator instanceof ClassMustHaveProperty);
        
        validator.overrideClass("TESTClosedWorldReasoner", "tcwr");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI("TEST421"));
        OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI("hasLvl2Pwd"));
 
        boolean isValid = validator.validateRequest(cls, odp);
        if (!isValid) {
            logger.info(validator.getExplanation());
        }
        assertTrue(isValid == true);

        String nextName = ClosedWorld.getNextClosedWorldReasonerName(validator.getCurrentClass(), validator.getCurrentPrefix());
        boolean isSaved = validator.saveLastRequest(nextName);
        assertTrue(isSaved == true);
    }
    
    /**
     * Validate known individuals against the "ClassMustHaveProperty" registered validator.
     * @throws Exception 
     */
    @Test 
    public void test_validateClassMustHaveProperty_c() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("ClassMustHaveProperty");
        assertTrue(validator instanceof ClassMustHaveProperty);
        
        validator.overrideClass("TESTClosedWorldReasoner", "tcwr");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI("TEST421"));
        OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI("hasLvlCPwd"));
 
        boolean isValid = validator.validateRequest(cls, odp);
        if (!isValid) {
            logger.info(validator.getExplanation());
        }
        assertTrue(isValid == true);
        
        if (isValid) {
            String nextName = ClosedWorld.getNextClosedWorldReasonerName(validator.getCurrentClass(), validator.getCurrentPrefix());
            boolean isSaved = validator.saveLastRequest(nextName);
            assertTrue(isSaved == true);

            boolean canReason = ClosedWorld.reason(validator.getCurrentClass(), validator.getCurrentPrefix());
            assertTrue(canReason == true);
        }
    }
    
    /**
     * Validate a known individual against the "ClassMustNotHaveProperty" registered validator.
     * @throws Exception 
     */
    @Test 
    public void test_validateClassMustNotHaveProperty_a() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("ClassMustNotHaveProperty");
        assertTrue(validator instanceof ClassMustNotHaveProperty);
        
        validator.overrideClass("TESTClosedWorldReasoner", "tcwr");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI("TEST421"));
        OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1OPwd"));
 
        boolean isValid = validator.validateRequest(cls, odp);
        if (!isValid) {
            logger.info(validator.getExplanation());
        }
        assertTrue(isValid == true);
        
        if (isValid) {
            String nextName = ClosedWorld.getNextClosedWorldReasonerName(validator.getCurrentClass(), validator.getCurrentPrefix());
            boolean isSaved = validator.saveLastRequest(nextName);
            assertTrue(isSaved == true);

            //Make sure we're passing the current class / prefix so we're validating against the test policy
            boolean canReason = ClosedWorld.reason(validator.getCurrentClass(), validator.getCurrentPrefix());
            assertTrue(canReason == true);
        }
    }
    
    /**
     * Validate a known individual against the "ClassMustNotHaveProperty" registered validator.
     * @throws Exception 
     */
    @Test 
    public void test_validateClassMustNotHaveProperty_b() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("ClassMustNotHaveProperty");
        assertTrue(validator instanceof ClassMustNotHaveProperty);
        
        validator.overrideClass("TESTClosedWorldReasoner", "tcwr");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI("TEST421"));
        OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI("hasLvlCPwd"));
 
        boolean isValid = validator.validateRequest(cls, odp);
        if (!isValid) {
            logger.info(validator.getExplanation());
        }
        assertTrue(isValid == false);
        
        String exp = validator.getExplanation();
        JSONObject expJSON = new JSONObject(exp);
        JSONArray failures = expJSON.getJSONArray("failures");
        //All items should fail this test
        assertTrue(failures.length() == 3);
    }
    
    /**
     * Validate a known individual against the "ClassMustNotHaveProperty" registered validator.
     * @throws Exception 
     */
    @Test 
    public void test_validateClassMustNotHaveProperty_c() throws Exception {
        Ontology.instance().prepareForTesting();
        
        ClosedWorld.Validator validator = ClosedWorld.getValidator("ClassMustNotHaveProperty");
        assertTrue(validator instanceof ClassMustNotHaveProperty);
        
        validator.overrideClass("TESTClosedWorldReasoner", "tcwr");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI("TEST421"));
        OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI("hasLvl2Pwd"));
 
        boolean isValid = validator.validateRequest(cls, odp);
        if (!isValid) {
            logger.info(validator.getExplanation());
        }
        assertTrue(isValid == false);
        
        String exp = validator.getExplanation();
        JSONObject expJSON = new JSONObject(exp);
        JSONArray failures = expJSON.getJSONArray("failures");
        //Make sure all 3items fail this test
        assertTrue(failures.length() == 3);
    }
}
