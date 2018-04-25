/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import static edu.illinois.iti.pbconf.ontology.Ontology.NEW_LINE;
import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421;

import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.util.List;
import java.util.Set;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.AfterClass;
import static org.junit.Assert.*;
import org.junit.Test;
import org.junit.Before;
import org.junit.BeforeClass;
import org.semanticweb.owlapi.model.AddAxiom;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassAssertionAxiom;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLException;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.reasoner.InconsistentOntologyException;
import org.semanticweb.owlapi.util.ShortFormProvider;
import org.semanticweb.owlapi.util.SimpleShortFormProvider;

/**
 *
 * @author anderson
 */
public class OntologyTest {
    static Logger logger = Logger.getLogger(OntologyTest.class.getName().replaceFirst(".+\\.",""));
    
    private static final String EXPECTED_ONTOLOGY_IRI="http://iti.illinois.edu/iti/pbconf/core";
    
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
     * Run one time configuration to create test data
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
     * Quick save by prefix
     *
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
     * Test to make sure known classes exist, and unknown does not
     * @throws Exception 
     */
    @Test
    public void testExistenceClass() throws Exception {
        boolean testA = Ontology.instance().classExists("SEL421");
        boolean testB = Ontology.instance().classExists("LINUX");
        boolean testC = Ontology.instance().classExists("FAKEFORTEST");
        assertTrue(testA && testB && !testC);
    }
    
    /**
     * Verify a known individual does exist, and unknown does not exist
     * @throws Exception 
     */
    @Test
    public void testExistenceIndividual() throws Exception {
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        boolean testA = Ontology.instance().individualExists("sel421FAKEDEVICEA");
        boolean testB = Ontology.instance().individualExists("sel421FAKEDEVICEANOTREAL");
        assertTrue(testA && !testB);
    }
    
    /**
     * Verify a known object property does exist, and unknown does not exist
     * @throws Exception 
     */
    @Test
    public void testExistenceObjectProperty() throws Exception {
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        boolean testA = Ontology.instance().objectPropertyExists("hasNTPStt");
        boolean testB = Ontology.instance().objectPropertyExists("hasINVALIDStt");
        assertTrue(testA && !testB);
    }
    
    /**
     * Verify a known data property does exist, and unknown does not exist
     * @throws Exception 
     */
    @Test
    public void testExistenceDataProperty() throws Exception {
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        boolean testA = Ontology.instance().dataPropertyExists("hasLvl1Pwd");
        boolean testB = Ontology.instance().dataPropertyExists("hasLvl1PwdINVALID");
        assertTrue(testA && !testB);
    }
    
    /**
     * Test of singleton.
     * @throws java.lang.Exception
     */
    @Test
    public void testInstance() throws Exception {
        Ontology ont = Ontology.instance();
        assertNotNull(ont);
    }

    /**
     * Make sure base IRI matches expected value
     * @throws Exception
     */
    @Test
    public void test_getBaseIRI() throws Exception {
        IRI iri = Ontology.instance().getRootIRI();
        logger.info("iri: "+iri);
        assertTrue("iri",iri.toString().equals(EXPECTED_ONTOLOGY_IRI));
    }
    
    /**
     * Make sure base ontology matches expected IRI
     * @throws Exception
     */
    @Test
    public void test_getBaseOntology() throws Exception {
        OWLOntology ont = Ontology.instance().getRootOntology();
        String iri = ont.getOntologyID().getOntologyIRI().toString();
        assertTrue("base ",iri.equals(EXPECTED_ONTOLOGY_IRI));
    }
    
    /**
     * Make sure we can start and run the OWL reasoner
     * @throws Exception
     */
    @Test
    public void test_startReasoner() throws Exception {
        Ontology.instance().startReasoner();
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Make sure the ontology IRI is what is expected for individuals
     * @throws Exception
     */
    @Test
    public void test_getIRI() throws Exception {
        IRI iri = Ontology.instance().getIRI("Phred");
        logger.info("getIRI: "+iri.toString());
        assertTrue("#Phred: ", iri.toString().equals(EXPECTED_ONTOLOGY_IRI+"#Phred"));
        String name = Ontology.instance().getSimpleName(iri.toString());
        assertTrue("pbconf:Phred: ", name.equals("pbconf:Phred"));
        name = Ontology.instance().getSimpleName(iri);
        assertTrue("pbconf:Phred: ", name.equals("pbconf:Phred"));
    }

    /**
     * Make sure IRI works with prefix
     * @throws Exception
     */
    @Test
    public void test_getIRIwPrefix() throws Exception {
        String CONFIG_URI = "http://iti.illinois.edu/iti/pbconf/config";
        Ontology.instance().setPrefix("config", IRI.create(CONFIG_URI));
        IRI iri = Ontology.instance().getIRI("config:Phred");
        logger.info("getIRI: "+iri.toString());
        assertTrue("Phred ", iri.toString().equals(CONFIG_URI+"#Phred"));
        String name = Ontology.instance().getSimpleName(iri.toString());
        assertTrue("config:Phred ", name.equals("config:Phred"));
        name = Ontology.instance().getSimpleName(iri);
        assertTrue("config:Phred ", name.equals("config:Phred"));
        // and back, just for giggles
        iri = Ontology.instance().getIRI(name);
        assertTrue(iri.toString().equals(CONFIG_URI+"#Phred"));
    }
    
    /**
     * Make sure IRI works with default prefix
     * @throws Exception
     */
    @Test
    public void test_getIRIwDefaultPrefix() throws Exception {
        String OWL_URI = "http://www.w3.org/2002/07/owl";
        IRI iri = Ontology.instance().getIRI("owl:Thing");
        logger.info("getIRI: "+iri.toString());
        assertTrue("Owl:Thing ", iri.toString().equals(OWL_URI+"#Thing"));
        String name = Ontology.instance().getSimpleName(iri.toString());
        assertTrue("owl:Thing ", name.equals("owl:Thing"));
        name = Ontology.instance().getSimpleName(iri);
        assertTrue("owl:Thing ", name.equals("owl:Thing"));
        // and back, just for giggles
        iri = Ontology.instance().getIRI(name);
        assertTrue(iri.toString().equals(OWL_URI+"#Thing"));
    }
    
    /**
     * Get an ontology from its prefix
     * @throws Exception
     */
    @Test
    public void test_getOntologyByPrefix() throws Exception {
        OWLOntology rootOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().get("coreOntology"));
        OWLOntology otherRootOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix());
        assertTrue(rootOnt == Ontology.instance().getRootOntology());
        assertTrue(rootOnt == otherRootOnt);
    }
    
    /**
     * Get the configuration ontology from its prefix
     * @throws Exception
     */
    @Test
    public void test_getConfigOntologyByPrefix() throws Exception {
        OWLOntology configOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());
        Set<OWLClass> classes = configOnt.getClassesInSignature();
        assertTrue(classes.size()>0);
    }
    
    /**
     * Get the policy ontology from its prefix
     * verify we can test individual from it
     * @throws Exception
     */
    @Test
    public void test_getPolicyOntologyByPrefix() throws Exception {
        OWLOntology policyOntology = Ontology.instance().getOntology(Ontology.instance().getConfig().getPolicyPrefix());
        Set<OWLClass> classes = policyOntology.getClassesInSignature();
        assertTrue(classes.size()>0);
    }
    
    /**
     * Make sure we have 0 instances by default
     * verify we can test individual from it
     * @throws Exception 
     */
    @Test
    public void test_dumpInstances() throws Exception {
        String output = Ontology.instance().dumpInstances("SEL421");
        logger.info(output);
        assertTrue("<#sel421",output.startsWith("<#sel421", 0));
    }

    /**
     * Make sure we have >0 instances of OWL thing by default
     * @throws Exception
     */
    @Test
    public void test_getInstances1() throws Exception {
        if (!Ontology.instance().isConstistent()) {
            logger.info(Ontology.instance().getFriendlyExplanation(Ontology.instance().getExplanation()));
            logger.info("test");
        }
        String classIRI = "owl:Thing";
        Set<OWLNamedIndividual> indivs = Ontology.instance().getInstances(classIRI);
        assertTrue(indivs.size()>0);
        // 32 of them now, but don't want to count on that
        logger.info("Individuals: "+indivs.toString());
    }
    
    /**
     * Test get instances with pizza ontology
     * @throws Exception
     */
    @Test
    public void test_getInstances2() throws Exception {
        if (!Ontology.instance().isConstistent()) {
            logger.info(Ontology.instance().getFriendlyExplanation(Ontology.instance().getExplanation()));
            logger.info("test");
        }
        // Show that we cannot get instances out of pizza if it is not imported to our root ontology.
        OWLOntology ont = Ontology.instance().loadOntology("owl/pizza.owl","pizza", false);
        assertTrue(ont != null);
        String classIRI = "http://www.co-ode.org/ontologies/pizza/pizza.owl#Country";
        //String classIRI = "owl:Thing";
        Set<OWLNamedIndividual> indivs = Ontology.instance().getInstances(classIRI);
        assertEquals(indivs.size(),0);
        logger.info("Individuals: "+indivs.toString());
    }
    
    /**
     * Make sure consistent ontology returns null explanation
     * @throws Exception
     */
    @Test
    public void test_explainNoConflict() throws Exception {
        if (!Ontology.instance().isConstistent()) {
            logger.info(Ontology.instance().getFriendlyExplanation(Ontology.instance().getExplanation()));
            logger.info("test");
        }
        assertTrue(Ontology.instance().getExplanation() == null);
    }

    /**
     * Verify we have an explanation for an inconsistent ontology
     * @throws Exception
     */
    @Test
    public void test_explanationsSEL421() throws Exception {
        Set<OWLNamedIndividual> instances = Ontology.instance().getInstances("SEL421");
        int count = instances.size();
        assertTrue(count>0);
        importProtocFile("newSEL421.bin",null);
        try {
            assertFalse(Ontology.instance().isConstistent());
            //instances = 
            Ontology.instance().getInstances("SEL421");
            fail("Should not get here. Expected exception instead.");
        } catch (InconsistentOntologyException ex) {
            logger.info("INCONSISTENCY DETECTED. That's good.");
        }
        // Now see if we can explain it!
        try {
            Set<Set<OWLAxiom>> formalExplanation = Ontology.instance().getExplanations();
            String explanation = Ontology.instance().getFriendlyExplanations(formalExplanation);
            assertTrue(explanation.length()>0);
            logger.info(explanation);
        } catch (Exception ex) {
            logger.error(ex);
            fail();
        }
    }

    /**
     * Make sure we can get a known individual
     * @throws Exception
     */
    @Test
    public void test_getIndividual() throws Exception {
        Individual ind = Ontology.instance().getIndividual("sel421FAKEDEVICEA", Ontology.instance().getRootOntology());
        assertTrue(ind != null);
    }
     
    /**
     * Test short form
     * @throws Exception
     */
    @Test
    public void test_ShortForm() throws Exception {
        ShortFormProvider sfp = new SimpleShortFormProvider();
        Set<OWLNamedIndividual> sels = Ontology.instance().getInstances("SEL421");
        logger.info(sfp.getShortForm(sels.iterator().next()));
    }

    /**
     * (borrowed temporarily from PBCONFServer.) 
     * imports SEL421List serialized protoc3 file. 
     * 
     * This is used to load a preserialized instance of a "bad" SEL421,
     * 
     * 
     * You can use the addsel421 script to create a new SEL421List file. 
     * 
     * @param  filename
     * @param prefix, if not null, the prefix for ontology to import into. Otherwise root
     * @return names of imported instances
     * @throws Exception 
     */
    private String importProtocFile(String filename, String prefix) throws Exception {
        SEL421Proto.SEL421List sel421List = SEL421Proto.SEL421List.parseFrom(new FileInputStream(filename));
        String output = "";
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLOntologyManager manager = Ontology.instance().getManager();
        OWLOntology oriOnt = Ontology.instance().getRootOntology();
        if (prefix != null) {
            oriOnt = Ontology.instance().getOntology(prefix);
        }

        for (SEL421 sel421 : sel421List.getSel421List()) {
            OWLClass SEL421 = dataFactory.getOWLClass(Ontology.instance().getIRI("SEL421"));
            OWLIndividual _sel421 = dataFactory.getOWLNamedIndividual(Ontology.instance().getIRI(sel421.getName()));
            OWLDataProperty _hasLvlCPwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvlCPwd"));
            OWLDataProperty _hasLvl1BPwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1BPwd"));
            OWLDataProperty _hasLvl1APwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1APwd"));
            OWLDataProperty _hasLvl1OPwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1OPwd"));
            OWLDataProperty _hasLvl2Pwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvl2Pwd"));
            OWLDataProperty _hasLvl1Pwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1Pwd"));
            OWLDataProperty _hasLvl1PPwd = dataFactory.getOWLDataProperty(Ontology.instance().getIRI("hasLvl1PPwd"));

            // Add instance
            OWLClassAssertionAxiom classAssertion = dataFactory.getOWLClassAssertionAxiom(SEL421, _sel421);
            manager.applyChange(new AddAxiom(oriOnt, classAssertion));

            // Add data property
            OWLDataPropertyAssertionAxiom axiom;
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvlCPwd, _sel421, sel421.getLvlCPwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvl1BPwd, _sel421, sel421.getLvl1BPwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvl1APwd, _sel421, sel421.getLvl1APwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvl1OPwd, _sel421, sel421.getLvl1OPwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvl2Pwd, _sel421, sel421.getLvl2Pwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvl1Pwd, _sel421, sel421.getLvl1Pwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));
            axiom = dataFactory.getOWLDataPropertyAssertionAxiom(_hasLvl1PPwd, _sel421, sel421.getLvl1PPwd());
            manager.applyChange(new AddAxiom(oriOnt, axiom));

            output += sel421.getName() + NEW_LINE;
        }

        return output;
    }
}

 