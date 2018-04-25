/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import java.io.File;
import java.io.FileNotFoundException;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONException;
import org.json.JSONObject;
import org.junit.AfterClass;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.semanticweb.owlapi.model.AddAxiom;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyExpression;
import org.semanticweb.owlapi.model.OWLDisjointClassesAxiom;
import org.semanticweb.owlapi.model.OWLEntity;
import org.semanticweb.owlapi.model.OWLException;
import org.semanticweb.owlapi.model.OWLIndividualAxiom;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLObjectProperty;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.reasoner.OWLReasoner;

/**
 *
 * @author anderson
 */
public class IndividualTest {

    static Logger logger = Logger.getLogger(IndividualTest.class.getName().replaceFirst(".+\\.", ""));

    static final String KNOWN_IND = "sel421FAKEDEVICEA";
    static final String KNOWN_SEL421 = "sel421FAKEDEVICEB";
    static final String KNOWN_SEL421_Port5 = "sel421FAKEDEVICEB_port5";

    private static boolean setUpIsDone = false;
    
    /**
     * Setup basic configuration
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
     * Get an individual from the root ontology
     * @throws Exception
     */
    @Test
    public void test_getIndividual() throws Exception {
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getRootOntology());
        assertTrue(ind != null);
    }

    /**
     * Get all classes an individual claims to be
     * @throws Exception
     */
    @Test
    public void test_getClasses() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getRootOntology());
        assertTrue(ind != null);

        Set<OWLClass> classes = ind.getClasses(KNOWN_IND);
        assertTrue(classes.size() > 0);

        Set<String> names = new HashSet<>();

        for (OWLClass cls : classes) {
            logger.info("class:" + cls.toString());
            names.add(Ontology.instance().getSimpleName(cls.toString()));
        }

        assertTrue(names.contains("<#SEL421>"));
        assertTrue(names.contains("<#Alarm>"));
        assertTrue(names.contains("<#AccessLogging>"));
    }

    /**
     * This test is workspace to debug oddity with OWLIndividual. An
     * OWLIndividual constructed with the dataFactory doesn't seem to have the
     * expected data properties, but a "same as" individual returned from the
     * reasoner does have the expected properties. This feels like a bug.
     *
     * @throws Exception
     */
    @Test
    public void test_debugDataProperties() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        OWLOntology cfgOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());
        OWLReasoner harry = Ontology.instance().getReasoner();

        IRI knownIRI = Ontology.instance().getIRI(KNOWN_IND);
        OWLNamedIndividual ind = Ontology.instance().getDataFactory().getOWLNamedIndividual(knownIRI);
        logger.info("constructed Individual: " + ind);
        Set<OWLNamedIndividual> sames = harry.getSameIndividuals(ind).getEntities();
        OWLNamedIndividual same = null;
        if (sames.size() > 0) {
            same = sames.iterator().next();
            Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp1 = ind.getDataPropertyValues(cfgOnt);
            Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp2 = same.getDataPropertyValues(cfgOnt);

            assertTrue(dataProp1.keySet().size() > 0);
            assertTrue(dataProp2.keySet().size() > 0);
        }
    }

    /**
     * This test is workspace to debug oddity with OWLIndividual. An
     * OWLIndividual constructed with the dataFactory doesn't seem to have the
     * expected data properties, but a "same as" individual returned from the
     * reasoner does have the expected properties. This feels like a bug.
     *
     * @throws Exception
     */
    @Test
    public void test_debugWhichIndividual() throws Exception {
        Ontology.instance().reset();
        Ontology.instance().initializeFromOWL("owl/whichindividual.ttl");
        
        Ontology.instance().prepareForTesting();

        // load a trivil ontology for this test
        OWLOntology ontology = Ontology.instance().getRootOntology();
        OWLReasoner reasoner = Ontology.instance().getReasoner();

        // Three should-be-equivalent ways to get IRI:
        IRI knownIRI = Ontology.instance().getIRI("something1");
        IRI knownIRI2 = IRI.create(ontology.getOntologyID().getOntologyIRI() + "#something1");
        IRI knownIRI3 = IRI.create(ontology.getOntologyID().getOntologyIRI().toString(), "#something1");
        assertTrue(knownIRI.toString().equals(knownIRI2.toString()));
        assertTrue(knownIRI.toString().equals(knownIRI3.toString()));

        OWLNamedIndividual ind = Ontology.instance().getDataFactory().getOWLNamedIndividual(knownIRI);
        OWLNamedIndividual ind2 = Ontology.instance().getDataFactory().getOWLNamedIndividual(knownIRI2);
        OWLNamedIndividual ind3 = Ontology.instance().getDataFactory().getOWLNamedIndividual(knownIRI3);

        Set<OWLNamedIndividual> sames = reasoner.getSameIndividuals(ind3).getEntities();
        OWLNamedIndividual same = null;
        if (sames.size() > 0) {
            same = sames.iterator().next();
        }
        assertTrue(same != null);
        logger.info("same from reasoner: " + same);
        Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp1 = ind.getDataPropertyValues(ontology);
        Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp2 = ind2.getDataPropertyValues(ontology);
        Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp3 = ind3.getDataPropertyValues(ontology);
        Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp4 = same.getDataPropertyValues(ontology);

        logger.info("Example of unexplained OWL API behavior:");
        assertTrue(dataProp1.keySet().size() == dataProp2.keySet().size());
        assertTrue(dataProp2.keySet().size() == dataProp3.keySet().size());
        assertTrue(dataProp3.keySet().size() == dataProp4.keySet().size());
        
    }

    /**
     * Extract and validate data properties from an individual
     * @throws Exception
     */
    @Test
    public void test_getDataProperties() throws Exception {
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()));
        assertTrue(ind != null);
        logger.info("ind iri: " + ind.getIRI());

        Set<String> propStrs = new HashSet<>();

        Set<OWLDataProperty> props = ind.getDataProperties();
        
        for (OWLDataProperty prop : props) {
            logger.info("property: " + prop.getIRI());
            propStrs.add(prop.getIRI().toString());
        }
        assertTrue(propStrs.contains(Ontology.instance().getIRI("hasLvlCPwd").toString()));
        assertTrue(propStrs.contains(Ontology.instance().getIRI("hasLvl1APwd").toString()));
        assertTrue(propStrs.contains(Ontology.instance().getIRI("hasLvl2Pwd").toString()));
    }

    /**
     * Extract and validate object properties from an individual
     * @throws Exception
     */
    @Test
    public void test_getObjectProperties() throws Exception {
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getRootOntology());
        assertTrue(ind != null);

        Set<String> propStrings = new HashSet<>();

        Set<OWLObjectProperty> props = ind.getObjectProperties();
        for (OWLObjectProperty prop : props) {
            logger.info("object property: " + prop.getIRI());
            propStrings.add(prop.getIRI().toString());
        }
        assertTrue(propStrings.contains(Ontology.instance().getIRI("hasPort5").toString()));
    }

    /**
     * Extract and validate axioms from an individual
     * @throws Exception
     */
    @Test
    public void test_getAxioms() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()));
        assertTrue(ind != null);

        Set<String> axiomStrings = new HashSet<>();

        Set<OWLIndividualAxiom> axioms = ind.getAxioms();
        for (OWLIndividualAxiom axiom : axioms) {
            logger.info("axiom: " + axiom.toString());
            // dataproperty object entity is the data type eg xsd:string. Hmm.
            // This is not very useful.
            Set<OWLEntity> entities = axiom.getSignature();
            for (OWLEntity entity : entities) {
                logger.info("entity: "+entity.toString());
            }
            axiomStrings.add(axiom.toString());
        }
        String hasPort5IRIString = Ontology.instance().getIRI("hasPort5").toString();
        // expect to find hasPort5IRIString in one of the axiomstrings.
        boolean foundIt = false;
        for (String axiomStr : axiomStrings) {
            if (axiomStr.contains(hasPort5IRIString)) {
                foundIt = true;
                break;
            }
        }
        assertTrue("hasPort5",foundIt);
    }
    
    /**
     * Set a data property with a string value, then retrieve the data property,
     * test its type, test its value.
     *
     * @throws Exception
     */
    @Test
    public void test_setPropertyStringValue() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()));
        assertTrue(ind != null);
        logger.info("ind iri: " + ind.getIRI());

        IRI hasStringA = Ontology.instance().getIRI("hasStringA");
        ind.setProperty(hasStringA, "Hello,World");
        Set<OWLLiteral> values = ind.getDataProperty(hasStringA);
        assertTrue(values.size() == 1);
        OWLLiteral value = values.iterator().next();
        logger.info(value.getDatatype());
        assertTrue(value.getDatatype().toString().equals("xsd:string"));
        logger.info(value.toString());
        logger.info(value.getLiteral());
        assertTrue(value.getLiteral().equals("Hello,World"));
    }

    /**
     * Set a data property with an int value, then retrieve the data property,
     * test its type, test its value.
     *
     * @throws Exception
     */
    @Test
    public void test_setPropertyIntegerValue() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()));
        assertTrue(ind != null);
        logger.info("ind iri: " + ind.getIRI());

        IRI hasIntA = Ontology.instance().getIRI("hasIntA");
        ind.setProperty(hasIntA, 42);
        Set<OWLLiteral> values = ind.getDataProperty(hasIntA);
        assertTrue(values.size() == 1);
        OWLLiteral value = values.iterator().next();
        logger.info(value.getDatatype());
        // Why does integer property show long form uri while string shows xsd:string? ARRG!
        assertTrue(value.getDatatype().toString().equals("http://www.w3.org/2001/XMLSchema#integer"));
        logger.info(value.toString());
        logger.info(value.getLiteral());
        assertTrue(value.parseInteger() == 42);
    }

    /**
     * Set a data property with an double value, then retrieve the data
     * property, test its type, test its value.
     *
     * @throws Exception
     */
    @Test
    public void test_setPropertyDoubleValue() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()));
        assertTrue(ind != null);
        logger.info("ind iri: " + ind.getIRI());

        IRI hasDoubleA = Ontology.instance().getIRI("hasDoubleA");
        ind.setProperty(hasDoubleA, 42.0042);
        Set<OWLLiteral> values = ind.getDataProperty(hasDoubleA);
        assertTrue(values.size() == 1);
        OWLLiteral value = values.iterator().next();
        logger.info(value.getDatatype());
        // Why does integer property show long form uri while string shows xsd:string? ARRG!
        assertTrue(value.getDatatype().toString().equals("http://www.w3.org/2001/XMLSchema#double"));
        logger.info(value.toString());
        logger.info(value.getLiteral());
        assertTrue(Math.abs(value.parseDouble() - 42.0042) < .00001);
    }

    /**
     * Set a data property with an boolean value, then retrieve the data
     * property, test its type, test its value.
     *
     * @throws Exception
     */
    @Test
    public void test_setPropertyBooleanValue() throws Exception {
        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        Individual ind = Ontology.instance().getIndividual(KNOWN_IND, Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()));
        assertTrue(ind != null);
        logger.info("ind iri: " + ind.getIRI());

        IRI hasBooleanA = Ontology.instance().getIRI("hasBooleanA");
        ind.setProperty(hasBooleanA, true);
        Set<OWLLiteral> values = ind.getDataProperty(hasBooleanA);
        assertTrue(values.size() == 1);
        OWLLiteral value = values.iterator().next();
        logger.info(value.getDatatype());
        // Why does integer property show long form uri while string shows xsd:string? ARRG!
        assertTrue(value.getDatatype().toString().equals("http://www.w3.org/2001/XMLSchema#boolean"));
        logger.info(value.toString());
        logger.info(value.getLiteral());
        assertTrue(value.parseBoolean() == true);
    }
    
    /**
     * Example of how to set and get an object property.
     *
     * For this test we use two individuals:
     * An individual with the iri config:cert2, representing a certificate, with the data property hasExpirationDate
     * individual config:sel421FAKEDEVICEB, with the object property hasCert.
     * 
     * @throws Exception
     */
    @Test
    public void test_SetGetObjectProperty() throws Exception {
        // load our complete standard ontology set
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);

        Ontology.instance().prepareForTesting();
        // Begin with an Individual representing the certificate:
        // we're going to put this new individual in the config ontology.
        OWLOntology confOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());
        // Note that the IRI for the new individual could be from any namespace.
        // It doesn't really matter for our reasoning what it is, but it would be
        // tidy for the indiv iri to be in the config namespace. To do that we
        // need to use the prefix config: on the individual name, else it will
        // use the root ontology iri (core).
        Individual cert2 = Ontology.instance().getIndividual("config:cert2", confOnt);
        assertTrue(cert2.getIRI().toString().contains(confOnt.getOntologyID().getOntologyIRI()));

        // This individual doesn't have any statements yet. Give it one:
        // cert2 hasExpirationDate "20161231"^^xsd:unsignedLong
        // I'm skeptical about using unsignedLong as a type for a date, but no matter for this test.
        // It presents an interesting problem, because there is no simple API support for
        // xsd:unsignedLong in our API or OWL API. Have to handroll.
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral dateLiteral = dataFactory.getOWLLiteral("20161231", org.semanticweb.owlapi.vocab.OWL2Datatype.XSD_UNSIGNED_LONG);

        logger.info(dateLiteral);
        // note that the namespace for hasExpirationDate _does_ matter, and it is core
        cert2.setProperty(Ontology.instance().getIRI("hasExpirationDate"), dateLiteral);

        // Then get the subject of our object property, and Individual sel421FAKEDEVICEB.
        // Object property objective:
        // sel421FAKEDEVICEB hasCert config:cert2
        Individual sel421FAKEDEVICEB = Ontology.instance().getIndividual("config:sel421FAKEDEVICEB", confOnt);
        sel421FAKEDEVICEB.setProperty(Ontology.instance().getIRI("hasCert"), cert2);

        // We should have set the object property, can we pull it all back?
        Set<OWLNamedIndividual> certs = sel421FAKEDEVICEB.getObjectProperty(Ontology.instance().getIRI("hasCert"));
        // there should be only one, because we just know that, and that makes it easy to get.
        assertTrue(certs.size() == 1);
        OWLNamedIndividual cert = certs.iterator().next();
        logger.info(cert.getIRI().toString());
        logger.info(cert2.getIRI().toString());
        // IRI should match the cert2 we created.
        assertTrue(cert.getIRI().toString().equals(cert2.getIRI().toString()));

        // get cert2.hasExpirationDate
        Individual sameAsCert2 = Ontology.instance().getIndividual(cert.getIRI().toString(), confOnt);
        Set<OWLLiteral> values = sameAsCert2.getDataProperty(Ontology.instance().getIRI("hasExpirationDate"));
        // there should be only one, because we just know that, and that makes it easy to get.
        assertTrue(values.size() == 1);
        OWLLiteral value = values.iterator().next();
        // the "literal" value of a data property is its string representation.
        logger.info(value.getLiteral());
        assertTrue(value.getLiteral().equals("20161231"));
    }

    /**
     * This tests a single object property:
     * The hasTelnetStt value (object) should be an IRI to the individual,
     *  http://iti.illinois.edu/iti/pbconf/core#off
     *
     * Setting the property to a string instead of an entity IRI 
     * throws a low-level AssertionError and can't even give an explanation.
     * 
     * @throws Exception
     */
    @Test
    public void test_SetGetBadObjectProperty() throws Exception {
        // load our complete standard ontology set
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);

        Ontology.instance().prepareForTesting();
        
        // we're going to put this property axiom in the config ontology.
        OWLOntology confOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());
        // backup ontology so we can rewind
        File confBackup = Ontology.instance().backup(Ontology.instance().getConfig().getConfigPrefix());
        
        // Then get the subject of our object property, and Individual sel421FAKEDEVICEB.
        Individual sel421FAKEDEVICEB = Ontology.instance().getIndividual("config:sel421FAKEDEVICEB", confOnt);
        
        // Set property with an INVALID value (string instead of object)
        sel421FAKEDEVICEB.setProperty(Ontology.instance().getIRI("hasTelnetStt"), "on");
        boolean isConsistent = false;
        
        try {
            isConsistent = Ontology.instance().isConstistent();            
        } catch (Throwable t) {
            logger.info("inconsistent. expected",t);
        }
        if (!isConsistent) {
            try {
                Set<Set<OWLAxiom>> expls = Ontology.instance().getExplanations();
                String fex = Ontology.instance().getFriendlyExplanations(expls);
                assertTrue(fex != null);
            } catch(Throwable t) {
                logger.error("can't even provide an explanation",t);
            }
        }
        else {
            fail("Did not expect consistent ontology");
        }

        // rewind to try something else
        Ontology.instance().restore(confBackup, Ontology.instance().getConfig().getConfigPrefix());
        confOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());
        sel421FAKEDEVICEB = Ontology.instance().getIndividual("config:sel421FAKEDEVICEB", confOnt);

        // Object property objective:
        // sel421FAKEDEVICEB hasTelnetStt on

        Individual on = Ontology.instance().getIndividual("on", Ontology.instance().getRootOntology());
        sel421FAKEDEVICEB.setProperty(Ontology.instance().getIRI("hasTelnetStt"), on);

        // confirm it:
        Set<OWLNamedIndividual> hasTelnetValue = sel421FAKEDEVICEB.getObjectProperty(Ontology.instance().getIRI("hasTelnetStt"));
        // there should be only one, and that makes it easy to get.
        assertTrue(hasTelnetValue.size() == 1);
        OWLNamedIndividual on2 = hasTelnetValue.iterator().next();
        logger.info(on.getIRI().toString());
        logger.info(on2.getIRI().toString());
        // IRI should match the cert2 we created.
        assertTrue(on.getIRI().toString().equals(on2.getIRI().toString()));

        // and should be consistent
        assertTrue(Ontology.instance().isConstistent());

    }

    /**
     * This tests a single object property:
     * The hasTelnetStt value (object) should be an IRI to the individual,
     *  http://iti.illinois.edu/iti/pbconf/core#off
     *
     * Setting the property to a string instead of an entity IRI 
     * throws a low-level AssertionError and can't even give an explanation.
     * 
     * @throws Exception
     */
    @Test
    public void test_SetGetAnotherBadObjectProperty() throws Exception {
        // load our complete standard ontology set
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);

        Ontology.instance().prepareForTesting();
        
        // we're going to put this property axiom in the config ontology.
        OWLOntology confOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());
        // backup ontology so we can rewind
        File confBackup = Ontology.instance().backup(Ontology.instance().getConfig().getConfigPrefix());
        assertTrue(confBackup != null);
        
        // Then get the subject of our object property, and Individual sel421FAKEDEVICEB.
        Individual sel421FAKEDEVICEB = Ontology.instance().getIndividual("config:sel421FAKEDEVICEB", confOnt);
        
        // object is a DigitalCertificate
        Individual cert2 = Ontology.instance().getIndividual("config:cert2", confOnt);
        assertTrue(cert2.getIRI().toString().contains(confOnt.getOntologyID().getOntologyIRI()));
        cert2.setClass(Ontology.instance().getIRI("DigitalCertificate"));
        
        assertTrue(Ontology.instance().isConstistent());

        // Do something odd, set hasTelnetStt to cert2
        // sel421FAKEDEVICEB hasTelnetStt cert2

        //Individual on = Ontology.instance().getIndividual("on", Ontology.instance().getRootOntology());
        sel421FAKEDEVICEB.setProperty(Ontology.instance().getIRI("hasTelnetStt"), cert2);

        // still consistent because reasoner doesn't know that cert2 is not a Status
        assertTrue(Ontology.instance().isConstistent());
        
        // Does reasoner infer that cert2 must be a Status? 
        boolean isCert2AStatus = false;
        Set<OWLNamedIndividual> stts = Ontology.instance().getInstances("Status");
        for (OWLNamedIndividual ind : stts) {
            logger.info(ind);
            if (ind.toString().contains("#cert2")) {
                isCert2AStatus = true;
            }
        }
        // Yes, it should be so according to the reasoner,
        // because the ontology says "any object of hasTelnetStt must be a Status"
        assertTrue(isCert2AStatus);
        
        
        // Low level tinkering with ontology...
        // Explicit change to the root ontology to declare that DigitalCert is disjunct with Status,
        
        java.util.Set<OWLClassExpression> classExpressions = new HashSet();
        classExpressions.add(Ontology.instance().getDataFactory().getOWLClass(
                Ontology.instance().getIRI("DigitalCertificate")));
        classExpressions.add(Ontology.instance().getDataFactory().getOWLClass(
                Ontology.instance().getIRI("Status")));
        OWLDisjointClassesAxiom disjoints = 
                Ontology.instance().getDataFactory().getOWLDisjointClassesAxiom(classExpressions);
        OWLOntologyManager manager = Ontology.instance().getManager();
        manager.applyChange(new AddAxiom(Ontology.instance().getRootOntology(), disjoints));
        
        // Now the ontology knows that the object of hasTelnetStt is not a Status,
        // so the ontology is inconsistent and an explanation is available. 
        assertFalse(Ontology.instance().isConstistent());
        Set<Set<OWLAxiom>> expls = Ontology.instance().getExplanations();
        String explanation = Ontology.instance().getFriendlyExplanations(expls);
        logger.info(explanation);
    }
    
    /**
     * Demonstrate that Individual works with non-root ontologies. Load an
     * alternate ontology, with one known instance, create an Individual for
     * that instance and read its data properties.
     *
     * @throws Exception
     */
    @Test
    public void test_altOntIndividual() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf_base.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf_base.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);
        
        Ontology.instance().prepareForTesting();
        
        // this funky single-purpose config ontology contains this one known instance.
        OWLOntology altOnt = Ontology.instance().loadOntology("owl/sel421b.ttl", "config", true);
        IRI altIRI = altOnt.getOntologyID().getOntologyIRI();
        logger.info(altIRI);
        OWLDataProperty hasLvlCPwd = null;
        String iriStr = altOnt.getOntologyID().getOntologyIRI().toString() + "#sel421FAKEDEVICEB";
        Individual ind = Ontology.instance().getIndividual(iriStr, altOnt);
        Set<OWLDataProperty> props = ind.getDataProperties();
        for (OWLDataProperty prop : props) {
            logger.info(prop.getIRI().toString());
            if (prop.getIRI().toString().contains("hasLvlCPwd")) {
                hasLvlCPwd = prop;
            }
        }
        assertTrue(hasLvlCPwd != null);
        // Now get its value...
        IRI hasLvlCPwdIRI = Ontology.instance().getIRI("hasLvlCPwd");
        Set<OWLLiteral> values = ind.getDataProperty(hasLvlCPwdIRI);
        assertTrue(values.size() == 1);
    }

    /**
     * Test all known data properties for an SEL421 Port 5 ethernet device
     *  Current known properties:
     *   - hasRetryDelayed
     *   - hasTelnetTimeout
     *   - hasMACAddr
     *   - hasTelnetPort
     *   - hasAccessTimeout
     * 
     * @throws Exception
     */
    @Test
    public void test_SEL421Port5DataProperties() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);

        Ontology.instance().reloadForTests();      
        List data = DataGenerator.getTestConfigurationDataRequestBoth();      
        for (Object dataStr : data) {
            assertTrue(operateAndCompare((String)dataStr, true));
        }
        Ontology.instance().prepareForTesting();
        
        //Load the config ontology
        OWLOntology confOnt = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix());

        //Get the known SEL421 device
        Individual sel421FAKEDEVICEB = Ontology.instance().getIndividual(KNOWN_SEL421, confOnt);
        logger.info("Known SEL421 device: " + sel421FAKEDEVICEB.getIRI().toString());
        assertTrue(sel421FAKEDEVICEB.getIRI().toString().contains(KNOWN_SEL421));

        //Use the known SEL421 device to get the known Port 5 ethernet device.
        Set<OWLNamedIndividual> port5 = sel421FAKEDEVICEB.getObjectProperty(Ontology.instance().getIRI("hasPort5"));

        //There should be only 1 port 5 device.
        assertTrue(port5.size() == 1);

        OWLNamedIndividual port5_iri = port5.iterator().next();
        logger.info("Know Port 5 device: " + port5_iri.toString());
        //The retreived device should have the known name.
        assertTrue(port5_iri.toString().contains(KNOWN_SEL421_Port5));

        //The retreived device should be in the config namespace
        Individual port5_ind = Ontology.instance().getIndividual(KNOWN_SEL421_Port5, confOnt);
        logger.info("Port 5 Individual: " + port5_ind.getName());
        assertTrue(port5_ind.getName().equals("pbconf:" + KNOWN_SEL421_Port5));

        Set<OWLNamedIndividual> port5_allowedOperation = port5_ind
                .getObjectProperty(Ontology.instance().getIRI("hasAllowedOperation"));

        Set<String> operations = new HashSet<>();

        for (OWLNamedIndividual operation : port5_allowedOperation) {
            logger.info("Allowed operation: " + operation.toString());
            operations.add(Ontology.instance().getSimpleName(operation.getIRI()));
        }

        assertTrue(operations.contains("pbconf:sel421_automation_setting"));
    }
}
