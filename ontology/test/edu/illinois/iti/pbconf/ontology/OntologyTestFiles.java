/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import static edu.illinois.iti.pbconf.ontology.Ontology.NEW_LINE;
import edu.illinois.iti.pbconf.ontology.SEL421Proto.SEL421;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.util.HashSet;
import java.util.Set;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.junit.AfterClass;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import org.junit.Before;
import org.junit.BeforeClass;
import org.junit.Test;
import org.semanticweb.owlapi.model.AddAxiom;
import org.semanticweb.owlapi.model.AddImport;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassAssertionAxiom;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLException;
import org.semanticweb.owlapi.model.OWLImportsDeclaration;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLIndividualAxiom;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyCreationException;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.reasoner.InconsistentOntologyException;
import org.semanticweb.owlapi.reasoner.InferenceType;
import org.semanticweb.owlapi.reasoner.Node;
import org.semanticweb.owlapi.util.AutoIRIMapper;

/**
 *
 * @author josephdigiovanna
 */
public class OntologyTestFiles {
    private static boolean setUpIsDone = false;
    
    static Logger logger = Logger.getLogger(OntologyTestFiles.class.getName().replaceFirst(".+\\.",""));
    
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
     * Run setup to make sure configuration ontology is cleared out.
     * @throws Exception
     */
    @Before
    public void setUp() throws Exception {
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

        Ontology.instance().replaceOntology(Ontology.instance().getConfig().getConfigPrefix());
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
     * Make sure we can start the reasoner
     * @throws Exception
     */
    @Test
    public void test_startReasoner() throws Exception {
        Ontology.instance().startReasoner();
        assertTrue(Ontology.instance().isConstistent());
    }
    
    /**
     * Make sure we can handle loading an invalid ontology file
     * @throws Exception
     */
    @Test
    public void testLoadOntologyNotFound() throws Exception {
        String notFile = "not.owl";
        try {
            Ontology.instance().loadOntology(notFile,null,true);
            fail("should have thrown");
        } catch (FileNotFoundException | OWLException e) {
            assertTrue(e instanceof FileNotFoundException);
        }

    }
    
    /**
     * Make sure we can load a bad ontology file (but file exists)
     * @throws Exception
     */
    @Test
    public void testLoadBadOntology() throws Exception {
        String notFile = "Makefile";
        try {
            Ontology.instance().loadOntology(notFile,null,true);
            fail("should have thrown");
        } catch (FileNotFoundException | OWLException e) {
            assertTrue(e instanceof OWLOntologyCreationException);
            logger.info("caught expected error loading ontology");
        }
    }
    
    /**
     * Make sure we can load a good ontology file
     * @throws Exception
     */
    @Test
    public void testLoadGoodOntology() throws Exception {
        String fp = "owl/pbconf.core.owl";
        Ontology.instance().reset();
        Ontology.instance().loadOntology(fp, Ontology.instance().getConfig().get("coreOntology"), true);
        assertTrue("ontology loaded.",true);
    }
    
    /**
     * Make sure we can load an additional ontology file
     * @throws Exception
     */
    @Test
    public void testLoadAnotherOntology() throws Exception {
        String fp = "owl/pizza.owl";
        OWLOntology pizza = Ontology.instance().loadOntology(fp,"pizza", true);
        // http://www.co-ode.org/ontologies/pizza/pizza.owl
        String newIRI = pizza.getOntologyID().getOntologyIRI().toString();
        logger.info("pizza ontology: "+newIRI);
        assertTrue("pizza: ", newIRI.equals("http://www.co-ode.org/ontologies/pizza/pizza.owl"));
    }
    
    /**
     * Make sure we can initialize with a configuration file
     * @throws Exception
     */
    @Test
    public void test_initialize() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        OWLOntology ont = Ontology.instance().initialize(cfg);
        assertTrue("initialized ",ont != null);
    }
    
    /**
     * Make sure we can initialize with JSON
     * @throws Exception
     */
    @Test
    public void test_initializeWithJSON() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        OWLOntology rootOnt = Ontology.instance().initialize(cfg);
        assertTrue("initialized ",rootOnt != null);
        assertTrue("rootOntology",rootOnt == Ontology.instance().getRootOntology());
    }
    
    /**
     * Make sure we can save an ontology file
     * @throws Exception
     */
    @Test
    public void test_saveOntology() throws Exception {
        String saveFileName = "./test_save.owl";
        File saveFile = new File(saveFileName);
        if (saveFile.exists()) {
            assertTrue(saveFile.delete());
        }
        assertTrue(saveFile.exists()==false);
        OWLOntology rootOnt = Ontology.instance().getRootOntology();
        Ontology.instance().saveOntology(rootOnt,saveFileName);
        assertTrue(saveFile.exists());
        assertTrue(saveFile.delete());
    }
    
    /**
     * Make sure we can save all open ontologies 
     * @throws Exception
     */
    @Test
    public void test_saveAll() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        
        Ontology.instance().loadOntology("owl/pizza.owl", null, false);
        Ontology.instance().saveAll(new File("backups/once"));
        File once = new File("backups/once");
        File conf = new File("backups/once/config.owl");
        File core = new File("backups/once/pbconf.owl");
        File pizza = new File("backups/once/pizza.owl");
        assertTrue(conf.exists());
        assertTrue(core.exists());
        assertTrue(pizza.exists());
        assertTrue(conf.delete());
        assertTrue(core.delete());
        assertTrue(pizza.delete());
        assertFalse(once.delete());
    }
    
    /**
     * Make sure we can backup an ontology
     * @throws Exception
     */
    @Test
    public void test_backupOntology() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);
        
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        
        OWLOntology configOnt = Ontology.instance().getOntology("config");
        assertTrue(configOnt != null);
        
        // backup
        File bkFile = Ontology.instance().backup("config");
        assertTrue(bkFile.exists());
        assertTrue(bkFile.getName().startsWith("config"));
        
        // poison
        importProtocFile("newSEL421.bin","config");
        assertFalse(Ontology.instance().isConstistent());
        
        // restore and delete backup
        Ontology.instance().restore(bkFile, "config");
        if (!Ontology.instance().isConstistent()) {
            Ontology.instance().getReasoner().flush();
            Ontology.instance().getReasoner().precomputeInferences(InferenceType.CLASS_ASSERTIONS);
            Set<Set<OWLAxiom>> formalExplanation = Ontology.instance().getExplanations();
            String explanation = Ontology.instance().getFriendlyExplanations(formalExplanation);
            logger.error(explanation);
        }
        assertTrue(Ontology.instance().isConstistent());
        assertTrue(bkFile.delete());
        assertFalse(bkFile.exists());
    }
    
    /**
     * Proves that we can load, import and query multiple ontologies.
     * @throws Exception 
     */
    @Test 
    public void test_importOntology() throws Exception {
        Ontology.instance().reset();        
        Ontology.instance().initializeFromOWL("owl/pbconf.core.owl");
                
        // There are no longer and defaul SEL421 devices, verify 0
        Set<OWLNamedIndividual> selsCore = Ontology.instance().getInstances("SEL421");
        assertTrue(selsCore.isEmpty());
        
        // Load the config ontology, which contains one more individual sel421FAKEDEVICEB
        Ontology.instance().loadOntology("owl/pbconf.config.owl","config", false);
        // expected IRI for that ontology:
       
        Ontology.instance().addConfigurationOntology();

        IRI configIRI = IRI.create("http://iti.illinois.edu/iti/pbconf/config");
        OWLOntology bOnt = Ontology.instance().getManager().getOntology(configIRI);
        // make sure we loaded the ontology IRI expected.
        assertTrue(bOnt.getOntologyID().getOntologyIRI().toString().equals(configIRI.toString()));
        
        // import new ontology into root
        OWLImportsDeclaration imports = Ontology.instance().getDataFactory().getOWLImportsDeclaration(configIRI);
        Ontology.instance().getManager().applyChange(new AddImport(Ontology.instance().getRootOntology(),imports));
    }
    
    /**
     * Make sure we can handle a conflicting SEL412 device
     * @throws Exception 
     */
    @Test
    public void test_ingestConflictingSEL421() throws Exception {
        Set<OWLNamedIndividual> instances = Ontology.instance().getInstances("SEL421");
        int count = instances.size();
        assertTrue(count==0);
        importProtocFile("newSEL421.bin",null);
        importProtocFile("newSEL421.bin",null);
        try {
            assertFalse(Ontology.instance().isConstistent());
            instances = Ontology.instance().getInstances("SEL421");
            logger.info(instances.toString());
            fail("Should not get here. Expected exception instead.");
        } catch (InconsistentOntologyException ex) {
            logger.info("INCONSISTENCY DETECTED. That's good.");
        }
        // Now see if we can explain it!
        try {
            Set<OWLAxiom> formalExplanation = Ontology.instance().getExplanation();
            String explanation = Ontology.instance().getFriendlyExplanation(formalExplanation);
            assertTrue(explanation.length()>0);
            logger.info(explanation);
        } catch (Exception ex) {
            logger.error(ex);
            fail();
        }
    }
    
    /**
     * Would an IRI mapper help us sensibly manage multiple ontologies? 
     * @throws Exception
     */
    @Test
    public void test_IRIMapper() throws Exception {
        File rootDirectory = new File("./owl");
        AutoIRIMapper mapper = new AutoIRIMapper(rootDirectory, false);
        Set<String> exts = new HashSet<>();
        exts.add(".ttl");
        exts.add(".owl");
        mapper.setFileExtensions(exts);
        mapper.update();
        
        java.util.Set<String> expected = new HashSet<>();
        expected.add("http://iti.illinois.edu/iti/pbconf/core");
        expected.add("http://iti.illinois.edu/iti/pbconf/config");
        
        java.util.Set<IRI> ontIRIs = mapper.getOntologyIRIs();
        logger.info("Found IRIs:");
        for (IRI iri : ontIRIs) {
            logger.info(iri.toString());
            if (expected.contains(iri.toString())) {
                expected.remove(iri.toString());
            }
        }
        assertTrue(expected.isEmpty());
        
        IRI pbconfIRI = Ontology.instance().getIRI("http://iti.illinois.edu/iti/pbconf/core");
        IRI docIRI = mapper.getDocumentIRI(pbconfIRI);
        logger.info("Document: "+docIRI.toString());
    }
    
    /**
     * Make sure we can handle unsatisfiable classes
     * @throws Exception 
     */
    @Test
    public void test_ingestUnsatisfiableClasses() throws Exception {
        Set<OWLNamedIndividual> instances = Ontology.instance().getInstances("SEL421");
        int count = instances.size();
        assertTrue(count>=0);
        try {
            Node<OWLClass> node = Ontology.instance().getReasoner().getUnsatisfiableClasses();
            Set<OWLClass> classes = node.getEntities();
            for (OWLClass cls : classes) {
                logger.info("unsatisfiable class: "+cls.toString());
            }
        } catch (InconsistentOntologyException ex) {
            logger.error(ex);
        }
    }
    
    /**
     * Validate another instances test
     * @throws Exception
     */
    @Test
    public void test_getInstances3() throws Exception {
        // Show that we CAN get instances out of pizza if it is imported to our root ontology.
        OWLOntology ont = Ontology.instance().loadOntology("owl/pizza.owl","pizza", true);
        assertTrue(ont != null);
        String classIRI = "http://www.co-ode.org/ontologies/pizza/pizza.owl#Country";
        //String classIRI = "owl:Thing";
        Set<OWLNamedIndividual> indivs = Ontology.instance().getInstances(classIRI);
        assertEquals(indivs.size(),5);
        logger.info("Individuals: "+indivs.toString());
    }
   
    /**
     * This test is mostly obsolete because we don't expect to use protoc buffered instances.
     * Also, newSEL421.bin loads an invalid instance (bad password), so the test changes
     * to detect inconsistent 
     * @throws Exception
     */
    @Test
    public void test_ingestSEL421() throws Exception {        
        Set<OWLNamedIndividual> instances = Ontology.instance().getInstances("SEL421");
        int count = instances.size();
        assertTrue(count>=0);
        try {
            // import invalid instance
            String output = importProtocFile("newSEL421.bin",null);       
            assertTrue(output != null);
            // show throw when we try to reason
            instances = Ontology.instance().getInstances("SEL421");
            logger.info(instances.toString());
            fail("newSEL421.bin was expected to be invalid.");
            assertTrue(instances != null); // trick findbugs
        } catch (InconsistentOntologyException ex) {
            logger.info("newSEL421.bin contains an expected invalid instance");
            return;
        }
    }

    /**
     * OWLOntology loadOntologyByPrefix(String prefix, boolean importToRoot)
     * Edited because config for pbconf causes ontologies to load
     * Now loading one that isn't part of the pbconf config file
     * @throws Exception
     */
    @Test
    public void test_loadOntologyByPrefix() throws Exception {
        IRI uIRI = IRI.create("http://iti.illinois.edu/iti/pbconf/test");
        Ontology.instance().setPrefix("test", uIRI);
        File ontDir = new File(Ontology.instance().getConfig().get("ontologyDirectory"));
        Ontology.instance().setOntologyDir(ontDir);
        OWLOntology uOnt = Ontology.instance().loadOntologyByPrefix("test", true);
        assertTrue(uOnt.getOntologyID().getOntologyIRI().equals(uIRI));
    }
    
    /**
     * Make sure we can save an ontology by its prefix
     * @throws Exception
     */
    @Test 
    public void test_saveOntologyByPrefix() throws Exception {
        IRI uIRI = IRI.create("http://iti.illinois.edu/iti/pbconf/test");
        Ontology.instance().setPrefix("test", uIRI);
        OWLOntology uOnt = Ontology.instance().loadOntologyByPrefix("test", true);
        assertTrue(uOnt.getOntologyID().getOntologyIRI().equals(uIRI));

        Individual ind = Ontology.instance().getIndividual("testIndividual", uOnt);
        Set<OWLIndividualAxiom> axioms = ind.getAxioms();
        assertTrue("New individual should have no axioms", axioms.isEmpty());
        // backup the test ontology before tinkering...
        File backup = Ontology.instance().backup("test");

        // now set some axiom on individual
        IRI clsIRI = Ontology.instance().getIRI("SEL421");
        ind.setClass(clsIRI);
        axioms = ind.getAxioms();
        assertTrue("NOW individual should have some axioms",axioms.size()>0);
        
        // save the test ontology with the new instance in it.
        File saved = Ontology.instance().saveOntologyByPrefix("test");
        assertTrue(saved.exists());
        
        Ontology.instance().getManager().removeOntology(uOnt);
        // Now it should not be loaded. 
        uOnt = Ontology.instance().getOntology("test");
        assertTrue("Ontology should be unloaded.",uOnt == null);
        
        // reload the ontology from file
        uOnt = Ontology.instance().loadOntologyByPrefix("test", true);
        assertTrue(uOnt.getOntologyID().getOntologyIRI().equals(uIRI));

        // Now the individual should have the axioms we saved
        ind = Ontology.instance().getIndividual("testIndividual", uOnt);
        axioms = ind.getAxioms();
        assertTrue("New individual should have axioms",axioms.size()>0);
        
        // Now restore from the backup with no axioms
        Ontology.instance().restore(backup, "test");
        // Now the individual should have no axioms again
        ind = Ontology.instance().getIndividual("testIndividual", uOnt);
        axioms = ind.getAxioms();
        assertTrue("New individual should have no axioms", axioms.isEmpty());
        
        // save the original test ontology
        Ontology.instance().saveOntologyByPrefix("test");

    }
    
    /**
     * This mega test is a fully commented walk-thru of basic operations 
     * with the PBCONF Ontology API.
     * @throws Exception 
     */
    @Test
    public void test_TutorialTest() throws Exception {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        Ontology.instance().initialize(cfg);
        
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        
        // Once initialized, we have three ontologies ready to go.
        // The core ontology is the first one loaded, 
        //and also the root ontology of the API. 
        OWLOntology rootOnt = Ontology.instance().getRootOntology();
        // we can also access ontologies by prefix. 
        // The core ontology prefix is "pbconf"
        OWLOntology pbconfOnt = Ontology.instance().getOntology("pbconf");
        // root and "pbconf" are the same ontology.
        assertTrue(rootOnt == pbconfOnt);
        // the "config" ontology is where we put device individuals
        OWLOntology configOnt = Ontology.instance().getOntology("config");
        // Use OWL API to get the ontology IRI, if you're interested.
        IRI configIRI = configOnt.getOntologyID().getOntologyIRI();
        assertTrue(configIRI.toString().equals("http://iti.illinois.edu/iti/pbconf/config"));

        // Before we start adding any configurations that could be invalid,
        // backup the config ontology.
        File configBackup = Ontology.instance().backup("config");
        
        // Let's add some configuration for a Schweitzer SEL421 device. 
        // This device configuration will be an individual in the config ontology.
        // The full unique name for this indivdual will be 
        // http://iti.illinois.edu/iti/pbconf/config#sel421FAKEDEVICEB
        // but using a prefix, we can call it "config:sel421FAKEDEVICEB". 
        Individual sel421FAKEDEVICEC = Ontology.instance().getIndividual("config:sel421FAKEDEVICEB", configOnt);
        
        // Tell the ontology that this individual is a SEL421.
        // The full IRI for the SEL421 class is 
        // http://iti.illinois.edu/iti/pbconf/core#SEL421
        // Since that's in the root ontology, we can get an IRI with 
        // the extra short name:
        IRI SEL421IRI = Ontology.instance().getIRI("SEL421");
        // Set the class axiom on the individual
        sel421FAKEDEVICEC.setClass(SEL421IRI);
        
        // We just changed the config ontology. 
        // Let's make sure everything is still consistent with the ontology.
        boolean isOk = Ontology.instance().isConstistent();
        assertTrue(isOk);
        
        // Now add a silly data property "hasTheUltimateAnswer" with an integer. 
        // The PBCONF ontology knows nothing about that property, and that's ok.
        IRI hasTheUltimateAnswer = Ontology.instance().getIRI("hasTheUltimateAnswer");
        sel421FAKEDEVICEC.setProperty(hasTheUltimateAnswer, 42);
        
        // Get the property back from the individual.
        Set<OWLLiteral> values = sel421FAKEDEVICEC.getDataProperty(hasTheUltimateAnswer);
        // The API allows that there might be more than one ultimate answer, 
        // so it returns a Set. 
        // For this example, we know that there is only one, 
        // so we take a short cut.
        assertTrue(values.size() == 1);
        OWLLiteral theAnswer = values.iterator().next();
        // OWLLiterals are always strings. 
        assertTrue(theAnswer.getLiteral().equals("42"));
        
        // Make sure the ontology is still consistent.
        assertTrue(Ontology.instance().isConstistent());
        
        // Let's add a more serious data property "hasLvl2Pwd"
        IRI hasLvl2Pwd = Ontology.instance().getIRI("hasLvl2Pwd");
        sel421FAKEDEVICEC.setProperty(hasLvl2Pwd, "TAIL");
        
        // It happens that TAIL is the default L2 password 
        // for this model of SEL421,
        // and that is not a good thing according to our ontology. 
        // We can confirm that it is inconsistent now!
        isOk = Ontology.instance().isConstistent();
        assertFalse(isOk);
        
        // But why is it inconsistent?
        // The formal explanation is a set of axioms
        Set<OWLAxiom> formalExplanation = Ontology.instance().getExplanation();
        // We can humanize it:
        String friendlyExplanation = Ontology.instance().getFriendlyExplanation(formalExplanation);
        // In this case the friendly explanation is: 
        String theExplanation = "\"A known default password value has been used for Level 2 password.\""+
                "\n\"A default password is a severe security risk.\"\n";
        assertTrue(friendlyExplanation.equals(theExplanation));
        
        // In PBCONF this configuration would not be allowed. 
        // A response would be returned to the PBCONF node in the form of:
        // { "status":"INVALID", "explanation":"\"A known default password value ..." }
        
        // After polluting the "config" ontology with inconsistent axioms,
        // we restore it from the backup to get back to normal.
        Ontology.instance().restore(configBackup, "config");
        assertTrue(Ontology.instance().isConstistent());
        
        // Lather, rinse, repeat.
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
