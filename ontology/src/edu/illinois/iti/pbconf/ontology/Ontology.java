/*
 * 
 */
package edu.illinois.iti.pbconf.ontology;

import com.clarkparsia.owlapi.explanation.BlackBoxExplanation;
import com.clarkparsia.owlapi.explanation.HSTExplanationGenerator;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;
import org.apache.log4j.Logger;
import org.semanticweb.HermiT.Configuration;
import org.semanticweb.HermiT.Reasoner;
import org.semanticweb.owlapi.apibinding.OWLManager;
import org.semanticweb.owlapi.model.AddImport;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLAnnotation;
import org.semanticweb.owlapi.model.OWLAnnotationProperty;
import org.semanticweb.owlapi.model.OWLAnnotationValue;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyExpression;
import org.semanticweb.owlapi.model.OWLDatatype;
import org.semanticweb.owlapi.model.OWLDatatypeRestriction;
import org.semanticweb.owlapi.model.OWLException;
import org.semanticweb.owlapi.model.OWLFacetRestriction;
import org.semanticweb.owlapi.model.OWLImportsDeclaration;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLObjectProperty;
import org.semanticweb.owlapi.model.OWLObjectPropertyExpression;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyCreationException;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.model.OWLOntologyStorageException;
import org.semanticweb.owlapi.model.RemoveImport;
import org.semanticweb.owlapi.reasoner.NodeSet;
import org.semanticweb.owlapi.reasoner.OWLReasoner;
import org.semanticweb.owlapi.reasoner.OWLReasonerFactory;
import org.semanticweb.owlapi.util.AutoIRIMapper;
import org.semanticweb.owlapi.util.OWLEntityRemover;
import org.semanticweb.owlapi.vocab.OWL2Datatype;
import org.semanticweb.owlapi.vocab.OWLFacet;
import uk.ac.manchester.cs.owl.owlapi.OWL2DatatypeImpl;

/**
 * This class provides a simplified interface for the PBCONF ontology. 
 * Designed to be used as a singleton. 
 * <p>
 * For PBCONF the Ontology instance should be initialized one time, with an
 * instance of OntologyConfig Ontology.instance().initialize(config)
 * <p>
 * For operations on individuals, the Individual class is provided.
 * <p>ni
 * See the test file OntologyTest.java for plentiful usage examples.
 *
 * @author anderson
 */
public class Ontology {

    private static Ontology _instance = null;
    static Logger logger = Logger.getLogger(Ontology.class.getName().replaceFirst(".+\\.", ""));
    
    /**
     * Just define HTML newline character
     */
    public static final String NEW_LINE = "<br>";

    // The instance configuration
  
    /**
     * Store local ontology configuration (loaded from file)
     */
    public OntologyConfig config = null;  

    // temporarily public so PBCONFServer can get at it during code transition
    private OWLOntologyManager manager = OWLManager.createOWLOntologyManager();
    private OWLDataFactory dataFactory = manager.getOWLDataFactory();
    private final OWLReasonerFactory reasonerFactory = new Reasoner.ReasonerFactory();

    // private ShortFormProvider sfp = new SimpleShortFormProvider(); // pretty useless
    private final HashMap<String,IRI> prefixes = new HashMap<>();

    // a mapper to automatically load ontologies from setOntologyDir()
    private AutoIRIMapper iriMapper = null;
    
    private OWLOntology rootOntology;
    private OWLReasoner reasoner = null;
    private IRI rootOntologyIRI = null;
    
    private boolean isConfigurationLoaded = false;
    private boolean isPartialConfigurationLoaded = false;

    private final File backupDir = new File("./backups");

    // firmly resist outside instantiation
    private Ontology() {}

    /**
     * @return singleton instance
     */
    public static synchronized Ontology instance() {
        if (_instance == null) {
            _instance = new Ontology();
            logger.debug("created instance");
        }
        return _instance;
    }

    /**
     * Set the directory used to autoload ontologies by IRI.
     * Typically Ontology.DEFAULT_ONT_DIR.
     * @param ontDir
     * @throws FileNotFoundException 
     */
    public void setOntologyDir(File ontDir) throws FileNotFoundException {
        if (!ontDir.isDirectory()) {
            throw new FileNotFoundException(ontDir.getPath());
        }
        AutoIRIMapper mapper = new AutoIRIMapper(ontDir, false);
        mapper.update();
        iriMapper = mapper;
    };
    
    /**
     * Load the ontology mapped to prefix, from the designated ontology dir.
     * @param prefix
     * @param importToRoot
     * @return ontology
     * @throws OWLException 
     */
    public OWLOntology loadOntologyByPrefix(String prefix, boolean importToRoot) throws OWLException {
        if (iriMapper == null) {
            return null;
        }
        IRI ontIRI = prefixes.get(prefix);
        OWLOntology ont = null;
        if (ontIRI != null) {
            IRI docIRI = iriMapper.getDocumentIRI(ontIRI);
            if (docIRI != null) {          
                try {
                    ont = loadOntology(docIRI, prefix, importToRoot);
                } catch (FileNotFoundException ex) {
                    logger.info("No ontology found for "+ontIRI);
                    throw new OWLOntologyStorageException("No ontology found for "+ontIRI);
                }
            }
        }
        return ont;
    }
    
    /**
     * Load ontology from filePath. 
     * The first ontology loaded is assumed to be the root ontology. All reasoning
     * is based on the root ontology and any other loaded ontologies which have been 
     * imported to the root. 
     * 
     * Each loaded ontology should preferably be assigned a prefix, by which that
     * ontology may be referenced in other Ontology functions, and which will be used
     * to create or interpret short names for entities. 
     * 
     * Prefix may be null, but why would you? 
     *
     * @param filePath
     * @param prefix a prefix string to reference this ontology in short IRIs. 
     * @param importToRoot new ontology will automatically be imported by root.
     * @throws FileNotFoundException
     * @return OWLOntology
     * @throws org.semanticweb.owlapi.model.OWLException
     */
    public OWLOntology loadOntology(String filePath, String prefix, boolean importToRoot) throws FileNotFoundException, OWLException {
        File ontFile = new File(filePath);
        if (!ontFile.exists()) {
            logger.info(filePath + " not found.");
            throw new FileNotFoundException(filePath);
        }
        IRI docIRI = IRI.create(ontFile);        
        return this.loadOntology(docIRI, prefix, importToRoot);
    }

    /**
     * Load ontology from IRI. 
     * The first ontology loaded is assumed to be the root ontology. All reasoning
     * is based on the root ontology and any other loaded ontologies which have been 
     * imported to the root. 
     * 
     * Each loaded ontology should preferably be assigned a prefix, by which that
     * ontology may be referenced in other Ontology functions, and which will be used
     * to create or interpret short names for entities. 
     * 
     * Prefix may be null, but why would you? 
     *
     * @param docIRI
     * @param prefix a prefix string to reference this ontology in short IRIs. 
     * @param importToRoot new ontology will automatically be imported by root.
     * @throws FileNotFoundException
     * @return OWLOntology
     * @throws org.semanticweb.owlapi.model.OWLException
     */
    public OWLOntology loadOntology(IRI docIRI, String prefix, boolean importToRoot) throws FileNotFoundException, OWLException {
        OWLOntology ontology;
        try {
            ontology = manager.loadOntologyFromOntologyDocument(docIRI);
            IRI ontIRI = ontology.getOntologyID().getOntologyIRI();
            if (rootOntology == null) {
                rootOntology = ontology;
                rootOntologyIRI = docIRI;
            } else if (importToRoot) {
                OWLImportsDeclaration imports = Ontology.instance().getDataFactory().getOWLImportsDeclaration(ontIRI);
                Ontology.instance().getManager().applyChange(new RemoveImport(Ontology.instance().getRootOntology(), imports));
                Ontology.instance().getManager().applyChange(new AddImport(Ontology.instance().getRootOntology(), imports));
            }
            if (prefix != null) {
                setPrefix(prefix,ontIRI);
            }
        } catch (OWLOntologyCreationException ex) {
            //logger.error("Loading ontology",ex);
            throw ex;
        }
        return ontology;
    }

    /**
     * Get root ontology, or null if no root ontology is loaded.
     * See loadOntology.
     *
     * @return OWLOntology
     */
    public OWLOntology getRootOntology() {
        return rootOntology;
    }

    /**
     * Return the loaded ontology matching the IRI, else null;
     *
     * @param ontIRI
     * @return
     */
    public OWLOntology getOntology(IRI ontIRI) {
        OWLOntology ont = manager.getOntology(ontIRI);
        return ont;
    }
    
    /**
     * Using the additionalOntologies array loaded by the configuration module,
     * find and return all classes that are loaded into the system.
     * @return 
     */
    public Set<OWLClass> getClassesInSignature() {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(true);
        Set<OWLClass> classes = new HashSet<>();
        
        for (Object ont : ontArr) {
            classes.addAll(Ontology.instance().getOntology((String)ont).getClassesInSignature());
        }
        
        return classes;
    }
    
    /**
     * Return the loaded ontology with registered prefix string, else null.
     * @param prefix
     * @return 
     */
    public OWLOntology getOntology(String prefix) {
        IRI iri = prefixes.get(prefix);
        if (iri != null) {
            return getOntology(iri);
        }
        return null;
    }
    
    /**
     * Unload all ontologies. 
     * You have to do this before you can reload a previously loaded ontology.
     */
    public void reset() {
        manager = OWLManager.createOWLOntologyManager();
        dataFactory = manager.getOWLDataFactory();
        rootOntology = null;
        rootOntologyIRI = null;
        iriMapper = null;
        reasoner = null;
        isConfigurationLoaded = false;
        isPartialConfigurationLoaded = false;
        prefixes.clear();
    }
    
    /**
     * Reset, and reload from configuration
     * @param savePolicy
     * @param saveConfiguration
     */
    public void reloadBaseOntologies(boolean savePolicy, boolean saveConfiguration) { 
        if (savePolicy) {
            try {
                this.saveOntologyByPrefix(this.config.get("policyOntology"));
            } catch (OWLException | FileNotFoundException ex) {
                logger.error("Unable to save policy ontology");
            }
        }
        if (saveConfiguration && isConfigurationLoaded) {
            try {
                this.saveOntologyByPrefix(this.config.get("configOntology"));
            } catch (OWLException | FileNotFoundException ex) {
                logger.error("Unable to save configuration ontology");
            }
        }
        try {
            this.reset();
            this.initialize(this.config);
        } catch (FileNotFoundException | OWLException ex) {
            logger.error("Error reloading base ontologies, exception = " + ex.toString());
        }
    }
    
    public void reloadForTests() {
        this.reloadBaseOntologies(false, false);
        this.addConfigurationOntology();
        Ontology.instance().replaceOntology("config");   
        Ontology.instance().replaceOntology("policy"); 
        this.reloadBaseOntologies(true, true);
    }
    
    public void prepareForTesting() {
        Ontology.instance().reloadBaseOntologies(false, false);
        Ontology.instance().addConfigurationOntology();
    }
    
    public boolean getIsConfigurationLoaded() {
        return isConfigurationLoaded;
    }
    
    public boolean getIsPartialConfigurationLoaded() {
        return isPartialConfigurationLoaded;
    }
    
    /**
     * Add in the partial configuration ontology in order to test consistency
     */
    public void addTemporaryConfigurationOntology() {
        try {
            if (!isPartialConfigurationLoaded) {
                this.loadOntologyByPrefix(this.config.get("partialConfigOntology"), true);
                isPartialConfigurationLoaded = true;
            }
        } catch (OWLException ex) {
            logger.error("Error loading the partial ontology, exception = " + ex.toString());
        }
    }
    
    /**
     * Add in the full configuration ontology so we can add the single piece of configuration to it.
     */
    public void addConfigurationOntology() {
        try {
            if (!isConfigurationLoaded) {
                this.loadOntologyByPrefix(this.config.get("configOntology"), true);
                isConfigurationLoaded = true;
            }
        } catch (OWLException ex) {
            logger.error("Error loading the configuration ontology, exception = " + ex.toString());
        }
    }
    
    /**
     * Copy all axioms over to the real configuration ontology
     */
    public void copyTemporaryToConfiguration() {
        Set<OWLAxiom> axiomsToAdd = new HashSet<>();
        OWLOntology tempOnt = this.getOntology(this.config.get("partialConfigOntology"));
        OWLOntology configOnt = this.getOntology(this.config.get("configOntology"));
        
        for (OWLAxiom ax : tempOnt.getAxioms()) {
            axiomsToAdd.add(ax);
        }
        
        manager.addAxioms(configOnt, axiomsToAdd);
    }
    
    /**
     * This will load the configuration ontology outside of root
     * then it will take each axiom from that configuration ontology, putting
     * it inside of the partial ontology in root, and checking consistency.
     * Doing it this way will allow us to detect all configuration issues at once,
     * not just the most immediate.  Will be testing explanations first to see if that catches everything or has a limit
     * @param type
     * @return 
     */
    public String validate(String type) {
        this.reloadBaseOntologies(true, true);
        //Add in the configuration ontology since we'll need to reason against it.
        this.addConfigurationOntology();
        String result = "";
        if (!this.isConstistent()) {
            result = this.getFriendlyExplanations(Ontology.instance().getExplanations());
        }
        
        if (type.equals("test")) {
            if (!ClosedWorld.reason(true)) {
                result = ClosedWorld.getExplanation();
            }
        } else {
            if (!ClosedWorld.reason(false)) {
                result = ClosedWorld.getExplanation();
            }
        }
       
        //Now we can reload again without saving (so we don't have configuration ontology loaded)
        this.reloadBaseOntologies(false, false);
        return result;
    }
    
    /**
     * Add configuration ontology in order to write configuration rule to it
     */

    /**
     * Start reasoner on root ontology.
     *
     * @throws java.lang.NullPointerException if rootOntology is null
     */
    public void startReasoner() {
        if (rootOntology == null) {
            throw new java.lang.NullPointerException("baseOntology");
        }
        OWLReasonerFactory rf = new Reasoner.ReasonerFactory();
        reasoner = rf.createNonBufferingReasoner(rootOntology);
        reasoner.precomputeInferences();
    }

    /**
     * Return the current reasoner, or null if not initialized.
     * @return OWLReasoner
     */
    public OWLReasoner getReasoner() {
        return reasoner;
    }

    /**
     * Return the current ontology manager, or null if not initialized.
     * @return OWLOntologyManager
     */
    public OWLOntologyManager getManager() {
        return manager;
    }

    /**
     * Return the current data factory, or null if not initialized.
     * @return OWLDataFactory
     */
    public OWLDataFactory getDataFactory() {
        return dataFactory;
    }
    
    /**
     * Return the current configuration loaded 
     * @return OntologyConfig
     */
    public OntologyConfig getConfig() {
        return config;
    }
    
    /**
     * Remove all axioms and individuals from an ontology
     * This is primarily used for the policy ontology reset
     * @param ont 
     */
    public void resetOntology(OWLOntology ont) {
        Set<OWLAxiom> axiomsToRemove = new HashSet<>();
        
        for (OWLAxiom ax : ont.getAxioms()) {
            String test = ax.toString();
            if (!ax.toString().contains("#test421") && !ax.toString().contains("#tcwr") && !ax.toString().contains("TEST421") && !ax.toString().contains("ClosedWorldReasoner")) {
                axiomsToRemove.add(ax);
            }
        }
        manager.removeAxioms(ont, axiomsToRemove);
        
        OWLEntityRemover remover = new OWLEntityRemover(manager, Collections.singleton(ont));
        
        //Remove all individuals that aren't prefixed with 'test421'
        for (OWLNamedIndividual ind : ont.getIndividualsInSignature()) {
            if (!ind.toString().contains("test421") && !ind.toString().contains("tcwr")) {
                ind.accept(remover);
            }
        }
        
        //Remove all classes
        for (OWLClass cls : ont.getClassesInSignature()) {
            if (!cls.toString().contains("TEST421") && !cls.toString().contains("ClosedWorldReasoner") && !cls.toString().contains("TestClosedWorldReasoner")) {
                cls.accept(remover);
            }
        }
        
        //Remove all object properties
        for (OWLObjectProperty obp : ont.getObjectPropertiesInSignature()) {
            //obp.accept(remover);
        }
        
        //Remove all data properties
        for (OWLDataProperty odp : ont.getDataPropertiesInSignature()) {
            //odp.accept(remover);
        }
        
        //Remove all data types
        for (OWLDatatype odt : ont.getDatatypesInSignature()) {
            //odt.accept(remover);
        }
        
        //Remove all anotation properties
        for (OWLAnnotationProperty oap : ont.getAnnotationPropertiesInSignature()) {
            //oap.accept(remover);
        }
        
        manager.applyChanges(remover.getChanges());
        remover.reset();
    }
    
    /**
     * Remove all matching object properties
     * @param ont 
     * @param indA 
     * @param indB 
     */
    public void removeMatchingObjectProperty(OWLOntology ont, Individual indA, Individual indB) {
        //Remove all matching object properties
        OWLDataFactory odf = manager.getOWLDataFactory();
        Set<OWLAxiom> axiomsToRemove = new HashSet<>();
        
        for (OWLObjectProperty odp : indA.getObjectProperties()) {            
            String odpTest = odp.toString().split("#")[1].replace(">", "");
            String propTest = indB.toString().split("#")[1].replace(">", "");
            if (propTest.equals(odpTest)) {               
                Set<OWLIndividual> inds = indA.getOWLIndividual().getObjectPropertyValues(odp, ont);
                for (OWLIndividual i : inds) {
                    OWLAxiom axiomToRemove = odf.getOWLObjectPropertyAssertionAxiom(odp, i, indB.getOWLIndividual());
                    axiomsToRemove.add(axiomToRemove);
                }
            }
        }
        
        manager.removeAxioms(ont, axiomsToRemove);
    }
    
    /**
     * Remove all matching data properties
     * @param ont
     * @param ind
     * @param propIRI 
     */
    public void removeMatchingDataProperty(OWLOntology ont, Individual ind, IRI propIRI) {
        //Remove all matching data properties
        OWLDataFactory odf = manager.getOWLDataFactory();
        Set<OWLAxiom> axiomsToRemove = new HashSet<>();
        
        for (OWLDataProperty odp : ind.getDataProperties()) {            
            String odpTest = odp.toString().split("#")[1].replace(">", "");
            String propTest = propIRI.toString().split("#")[1].replace(">", "");
            if (propTest.equals(odpTest)) {
                Set<OWLLiteral> lits = ind.getOWLIndividual().getDataPropertyValues(odp, ont);
                for (OWLLiteral i : lits) {
                    OWLAxiom axiomToRemove = odf.getOWLDataPropertyAssertionAxiom(odp, ind.getOWLIndividual(), i.getLiteral());
                    axiomsToRemove.add(axiomToRemove);
                }
            }
        }
        
        manager.removeAxioms(ont, axiomsToRemove);
    }
    
    /**
     * Load an ontology configuration from a specified JSON file.
     * @param cfg
     * @return rootOntology
     * @throws FileNotFoundException
     * @throws OWLException 
     */
    public OWLOntology initialize(OntologyConfig cfg) throws FileNotFoundException, OWLException {
        if (rootOntology != null) {
            logger.info("restarting ontology");
            reset();
        }
        
        config = cfg;
        setDefaultPrefixes();
        setConfigPrefixes();      
        OWLOntology coreOnt = initializeOntology();
        loadAdditionalOntologies();
        startReasoner();
   
        return coreOnt;
    }
    
    /**
     * Initialize the ontology
     * @return
     * @throws FileNotFoundException
     * @throws OWLException
     */
    public OWLOntology initializeOntology() throws FileNotFoundException, OWLException {
        setOntologyDir(new File(config.get("ontologyDirectory")));
        return loadOntologyByPrefix(config.get("coreOntology"), false); 
    }
    
    /**
     * Load the additional ontologies
     * @throws org.semanticweb.owlapi.model.OWLException
     * @throws java.io.FileNotFoundException
     */
    public void loadAdditionalOntologies() throws OWLException, FileNotFoundException {
        ArrayList ontArr = config.getAdditionalOntologies();      
        for (Object ontArrObj : ontArr) {
            processOntology((String) ontArrObj);
        }
    }
    
    /**
     * Process each ontology we want to load
     * May take the following forms : 
     * 
     * 1.) Prefix
     * 2.) IRI
     * 3.) File IRI
     * 
     * @param ont 
     * @throws org.semanticweb.owlapi.model.OWLException 
     * @throws java.io.FileNotFoundException 
     */
    public void processOntology(String ont) throws OWLException, FileNotFoundException {
        boolean added = false;
        
        for (Map.Entry<String, String> prefix : config.getPrefixes().entrySet()) {
            if (added == false && prefix.getKey().equals(ont)) {
                loadOntologyByPrefix(ont, true);
                added = true;
            }
            if (added == false && prefix.getValue().equals(ont)) {
                loadOntologyByPrefix(prefix.getKey(), true);
                added = true;
            }
        }
        
        //At this point we know we don't have a matching prefix or iri in the system
        //Now we determine if it's a file iri or an iri
        if (added == false) {
            if (ont.startsWith("http:")) {
                
            } else if (ont.startsWith("file:")) {
                String ontFile = ont.replace("file:", "");
                String filePath = System.getProperty("user.dir");
                filePath = filePath.concat("/");
                filePath = filePath.concat(ontFile);
                
                if (filePath.toLowerCase(Locale.US).endsWith(".owl") || filePath.toLowerCase(Locale.US).endsWith(".xml")) {
                    Path fp = Paths.get(ontFile);
                    String fileName = fp.getFileName().toString();  
                    fileName = fileName.replace(".owl", "");
                    fileName = fileName.replace(".xml", "");
                    Ontology.instance().loadOntology(ontFile, fileName, true);                   
                } else if (filePath.toLowerCase(Locale.US).endsWith(".ttl")) {
                    //We'll need to create a prefix from the filename and then import it
                    Path fp = Paths.get(ontFile);
                    String fileName = fp.getFileName().toString();  
                    fileName = fileName.replace(".ttl", "");
                    Ontology.instance().loadOntology(ontFile, fileName, true);
                } else {
                    logger.warn("Unable to load ontology reference, unsupported file type " + ont);
                    //We can't currently support this file format
                }             
            } else {
                //We aren't able to load this, skipping
                logger.warn("Unable to load ontology reference " + ont);
            }
        }
    }
        
    /**
     * Load all prefixes that are part of the configuration file
     * You can call this after extending / altering the prefix
     * array in order to reload them all.
     */
    public void setConfigPrefixes() {
        for (Map.Entry<String, String> prefix : config.getPrefixes().entrySet()) {
            setPrefix(prefix.getKey(), IRI.create(prefix.getValue()));
        }
    }

    /**
     * Load a specific root ontology file and start reasoner. 
     * Will restart the ontology if needed.
     * 
     * @deprecated this is now used only for specific unit tests
     * @param ontFilePath
     * @throws FileNotFoundException
     * @throws OWLException
     * @return OWLOntology base ontology
     */
    public OWLOntology initializeFromOWL(String ontFilePath) throws FileNotFoundException, OWLException {
        if (rootOntology != null) {
            logger.info("restarting ontology with " + ontFilePath);
            reset();
        }
        OWLOntology ont = loadOntology(ontFilePath, config.get("coreOntology"), true);
        startReasoner();        
        setDefaultPrefixes();
        
        return ont;
    }
    
    /**
     * get root ontology IRI
     *
     * @throws java.lang.NullPointerException if rootOntology is null
     * @return String
     */
    public IRI getRootIRI() {
        //OntologyIRIMappingNotFoundException(IRI ontologyIRI) 
        if (rootOntology == null) {
            throw new java.lang.NullPointerException("baseOntology");
        }
        return rootOntology.getOntologyID().getOntologyIRI();
    }

    /**
     * Convert a short name into a long IRI name. 
     * Looks for prefixes set by setPrefix, as well as default prefixes e.g. 
     * "owl" and "xsd". If name begins with "http:" assume it is a full IRI,
     * which is a bit lame but works for our purposes. 
     * Otherwise, assume name is a leaf in the base ontology.
     * @param name
     * @return IRI
     */
    public IRI getIRI(String name) {
        String baseIRI = getRootIRI().toString();
        IRI iri;
        
        Set<String>pfxs = prefixes.keySet();
        for (String prefix : pfxs) {
            String regex = "^"+prefix+":.*";
            if (name.matches(regex)) {
                String leaf = name.substring(prefix.length()+1);
                IRI resolvedIRI = prefixes.get(prefix);
                iri = IRI.create(resolvedIRI.toString()+"#"+leaf);
                return iri;
            }
        }
        
        if (name.matches("^http:.*")) {
            iri = IRI.create(name);
            return iri;
        }
        
        return IRI.create(baseIRI + "#" + name);
    }

    /**
     * Return a simplified name for the iri. 
     * This will return a prefixed name if the iri is found in the current 
     * prefix mappings. See setPrefix()
     * Otherwise just returns the full IRI string.
     *
     * @param iri
     * @return String
     */
    public String getSimpleName(String iri) {
        IRI baseIRI = getRootIRI();
        
        Set<String>pfxs = prefixes.keySet();
        for (String prefix : pfxs) {
            String resolvedIRI = prefixes.get(prefix).toString();
            String regex = "^"+resolvedIRI+"#.*";
            if (iri.matches(regex)) {
                String leaf = iri.substring(resolvedIRI.length()+1); // include #
                String simpleName = prefix+":"+leaf;
                return simpleName;
            }
        }
        
        return iri.replaceAll(baseIRI.toString(), "");
    }

    /**
     * Return a simplified name for the iri. 
     * This will return a prefixed name if the iri is found in the current 
     * prefix mappings. See setPrefix()
     * Otherwise just returns the full IRI string.
     *
     * @param iri
     * @return String
     */
    public String getSimpleName(IRI iri) {
        return getSimpleName(iri.toString());
    }

    /**
     * Associate a prefix with an IRI.
     * The prefix should be a short string without a trailing :, 
     * and the IRI should not include a trailing #.
     * @param prefix
     * @param iri 
     */
    public void setPrefix(String prefix, IRI iri) {
        prefixes.put(prefix, iri);
    }

    /**
     * Set default prefixes for owl, xsd, rdf, and rdfs. 
     * This is called automatically by intialize()
     */
    public void setDefaultPrefixes() {
        // set some default prefixes
        setPrefix("owl",IRI.create("http://www.w3.org/2002/07/owl"));
        setPrefix("xsd",IRI.create("http://www.w3.org/2001/XMLSchema"));
        setPrefix("rdf",IRI.create("http://www.w3.org/1999/02/22-rdf-syntax-ns"));
        //setPrefix("xml",IRI.create("http://www.w3.org/XML/1998/namespace"));
        setPrefix("rdfs",IRI.create("http://www.w3.org/2000/01/rdf-schema"));
    }
    
    /**
     * Return a new Individual, with simplified interface to common operations.
     * The ontology specified will be the home ontology for the individual. This
     * is where all new axioms will be saved.
     *
     * @param name simple name or IRI string
     * @param ontology the "home" ontology for this individual
     * @return
     */
    public Individual getIndividual(String name, OWLOntology ontology) {
        return new Individual(getIRI(name),ontology);
    }

    /**
     * get the instances of a named class.
     *
     * @param className
     * @return Set OWLNamedIndividual 
     */
    public Set<OWLNamedIndividual> getInstances(String className) {
        if (reasoner == null) {
            _instance.startReasoner();
        }
        
        OWLClass myclass = dataFactory.getOWLClass(getIRI(className));
        NodeSet<OWLNamedIndividual> individualNodeSet = reasoner.getInstances(myclass, false);
        Set<OWLNamedIndividual> individuals = individualNodeSet.getFlattened();
        return individuals;
    }

    /**
     * Get all types for an individual
     *
     * @param individualName
     * @return
     */
    public Set<OWLClass> getTypes(String individualName) {
        OWLNamedIndividual individual = dataFactory.getOWLNamedIndividual(getIRI(individualName));
        NodeSet<OWLClass> classNodeSet = reasoner.getTypes(individual, true);
        Set<OWLClass> classes = classNodeSet.getFlattened();
        return classes;
    }

    /**
     * return all subclasses of a class
     *
     * @param className
     * @return
     */
    public Set<OWLClass> getSubclasses(String className) {
        OWLClass myclass = dataFactory.getOWLClass(getIRI(className));
        NodeSet<OWLClass> subClassNodeSet = reasoner.getSubClasses(myclass, false);
        Set<OWLClass> classes = subClassNodeSet.getFlattened();
        return classes;
    }

    /**
     * return all superclasses of a class
     *
     * @param className
     * @return
     */
    public Set<OWLClass> getSuperclasses(String className) {
        OWLClass myclass = dataFactory.getOWLClass(getIRI(className));
        NodeSet<OWLClass> classNodeSet = reasoner.getSuperClasses(myclass, false);
        Set<OWLClass> classes = classNodeSet.getFlattened();
        return classes;
    }

    /**
     * Is the current ontology consistent? 
     * Else, there must be an invalid configuration.
     *
     * @return is consistent.
     */
    public boolean isConstistent() {
        return reasoner.isConsistent();
    };
    
    /**
     * Used by getExplanation(s).
     * @private
     * @return 
     */
    private HSTExplanationGenerator getExplaner() {
        // What a monster. We can't get an explanation from our regular reasoner, because 
        // it is throws an InconsistentOntologyException when we try to call getExplanation.
        // Therefore we have to create a new special reasoner with exceptions suppressed.
        // bless you stackoverflow: http://stackoverflow.com/questions/26545241/owlexplanation-with-hermit-reasoner

        Configuration configuration = new Configuration();
        configuration.throwInconsistentOntologyException = false;

        OWLReasonerFactory factory = new Reasoner.ReasonerFactory() {
            @Override
            protected OWLReasoner createHermiTOWLReasoner(org.semanticweb.HermiT.Configuration configuration, OWLOntology ontology) {
                // don't throw an exception since otherwise we cannot compte explanations 
                configuration.throwInconsistentOntologyException = false;
                return new Reasoner(configuration, ontology);
            }
        };
        OWLReasoner customReasoner = reasonerFactory.createReasoner(rootOntology, configuration);
        BlackBoxExplanation expl = new BlackBoxExplanation(getRootOntology(), factory, customReasoner);
        HSTExplanationGenerator explanator = new HSTExplanationGenerator(expl);
        return explanator;
    }

    /**
     * Get the axioms that comprise the first returned explanation for an
     * inconsistent ontology. If the ontology is consistent, then returns null.
     *
     * @return set of axioms or null
     */
    public Set<OWLAxiom> getExplanation() {
        if (reasoner.isConsistent()) {
            return null;
        }
        HSTExplanationGenerator explanator = getExplaner();
        Set<OWLAxiom> explanation = explanator.getExplanation(dataFactory.getOWLThing());
        return explanation;
    }

    /**
     * Get the axioms that comprise all known explanations for an inconsistent
     * ontology. If the ontology is consistent, then returns null.
     *
     * @return set of axioms or null
     */
    public Set<Set<OWLAxiom>> getExplanations() {
//        if (reasoner.isConsistent()) {
//            return null;
//        }
        HSTExplanationGenerator explanator = getExplaner();
        Set<Set<OWLAxiom>> explanations = explanator.getExplanations(dataFactory.getOWLThing());
        return explanations;
    }

    /**
     * Get a human-friendly text explanation from the formal explanation. 
     * The preferred output is derived from rdfs:comments on the axioms that
     * comprise the explanation. If no comments are available, a raw dump of the
     * axioms is provided, as a last resort.
     *
     * @param explanation
     * @return String
     */
    public String getFriendlyExplanation(Set<OWLAxiom> explanation) {
        StringBuilder raw = new StringBuilder("Sorry, I don't have a friendly explanation. The best I can do is:\n");
        StringBuilder talk = new StringBuilder();
        if (explanation == null) {
            talk.append("The ontology is consistent. Nothing to explain.");
        } else {
            for (OWLAxiom axiom : explanation) {
                raw.append(axiom.toString()).append("\n");
                Set<OWLAnnotation> annotations = axiom.getAnnotations();
                for (OWLAnnotation note : annotations) {
                    OWLAnnotationProperty property = note.getProperty();
                    if (property.isComment()) {
                        OWLAnnotationValue value = note.getValue();
                        talk.append(value.toString()).append("\n");
                        logger.debug("comment: " + value.toString());
                    }
                }
            }
        }
        if (talk.length() == 0) {
            // no comments found so the best we can do is return raw
            return raw.toString();
        }
        return talk.toString();
    }

    /**
     * Get a human-friendly text explanation from the formal explanations. 
     * The preferred output is derived from rdfs:comments on the axioms that
     * comprise the explanation. If no comments are available, a raw dump of the
     * axioms is provided, as a last resort.
     *
     * @param explanations
     * @return String
     */
    public String getFriendlyExplanations(Set<Set<OWLAxiom>> explanations) {
        StringBuilder talk = new StringBuilder();
        int count = 1;
        for (Set<OWLAxiom> explanation : explanations) {
            talk.append("Explanation ").append(count).append(" -------------------------------------\n");
            talk.append(getFriendlyExplanation(explanation));
            count++;
        }
        return talk.toString();
    }
    /**
     * Save the ontology mapped to prefix, to the designated ontology directory.
     * @param prefix
     * @return file
     * @throws OWLException 
     * @throws java.io.FileNotFoundException 
     */
    public File saveOntologyByPrefix(String prefix) throws OWLException, FileNotFoundException {
        if (iriMapper == null) {
            return null;
        }
        File ontFile = null;
        IRI ontIRI = prefixes.get(prefix);
        if (ontIRI != null) {
            OWLOntology ont = Ontology.instance().getOntology(ontIRI);
            if (ont == null) {
                throw new OWLOntologyStorageException("Ontology not found for prefix "+prefix);
            }
            IRI docIRI = iriMapper.getDocumentIRI(ontIRI);
            if (docIRI != null) {
                File docFile = new File(docIRI.toURI());
                String docPath = docFile.getPath();
                ontFile = saveOntology(ont, docPath);
            }
        }
        return ontFile;
    }
    
    /**
     * Save ontology to a file.
     *
     * @param ont
     * @param filename
     * @throws OWLOntologyStorageException
     * @throws FileNotFoundException
     * @return saved ontology file
     */
    public File saveOntology(OWLOntology ont, String filename) throws OWLOntologyStorageException, FileNotFoundException {
        File ontFile = new File(filename);
        FileOutputStream fos = new FileOutputStream(ontFile);
        try {
            manager.saveOntology(ont, fos);
        } 
        finally {
            try {
                fos.close();
            } catch (IOException ex) {
                logger.error("Saving ontology:",ex);
                throw new OWLOntologyStorageException("Error saving ontology.",ex);
            }
        }
        return ontFile;
    }

    /**
     * Save a copy of all currently loaded ontologies to files in the specified directory.
     * If the ontologies have a registered prefix, then the file will be named *prefix*.owl.
     * Otherwise a suitable short name is created. 
     * 
     * (I originally intended this to be part of a global backup and restore feature, 
     * but I think it may be more effective to backup/restore the individual ontology 
     * files as we add configs and policies. It is currently unused in PBCONF.)
     * 
     * 
     * @param directory
     * @throws FileNotFoundException
     * @throws OWLOntologyStorageException 
     */
    public void saveAll(File directory) throws FileNotFoundException, OWLOntologyStorageException {
        if (!directory.exists() && !directory.mkdir()) {
            throw new FileNotFoundException(directory.getPath());
        }
        // these iterators are oldschool, but I want saveOntology to throw,
        // and I don't know how to do that from inside a forEach 
        Set<OWLOntology> onts = Ontology.instance().getManager().getOntologies();
        Iterator<OWLOntology> it = onts.iterator();
        while (it.hasNext()) {
            OWLOntology ont = it.next();
            IRI ontIRI = ont.getOntologyID().getOntologyIRI();
            String iriStr = ontIRI.toString();
            String filename = iriStr.substring(iriStr.lastIndexOf('/')+1);
            if (!filename.endsWith(".owl")) {
                filename += ".owl";
            }
            
            // TODO: need getPrefix fn
            if (prefixes.containsValue(ontIRI)) {
                Set<Map.Entry<String, IRI>> entries = prefixes.entrySet();
                Iterator<Map.Entry<String, IRI>> pit = entries.iterator();
                while (pit.hasNext()) {
                    Map.Entry<String, IRI> entry = pit.next();
                    if (ontIRI.equals(entry.getValue())) {
                        String prefix = entry.getKey();
                        filename = prefix+".owl";
                        break;
                    }
                }
            }
            File bkFile = new File(directory,filename);
            saveOntology(ont,bkFile.getPath());
        }
    }
    
    /**
     * Backup the ontology referenced by a registered prefix. 
     * This should be used to backup an ontology before inserting new, potentially 
     * invalid statements. The resulting backup file can be used to restore 
     * the ontology to its previous state if inconsistency results. Note that the 
     * caller is responsible for deleting the backup file if that is desired.
     * 
     * @param prefix
     * @return File - backup file, or null ontology is not found for prefix
     * @throws OWLOntologyStorageException
     * @throws FileNotFoundException - if backup dir cannot be found or created.
     */
    public File backup(String prefix) throws OWLOntologyStorageException, FileNotFoundException {
        if (!backupDir.exists() && !backupDir.mkdir()) {
            throw new FileNotFoundException(backupDir.getPath());
        }
        OWLOntology ont = Ontology.instance().getOntology(prefix);
        if (ont != null) {
            long now = (long) (System.currentTimeMillis());// / 1000L);
            String bkFileName = prefix + "." + now;
            File bkFile = new File(backupDir, bkFileName);
            saveOntology(ont,bkFile.getPath());
            return bkFile;
        }
        return null;
    }

    /**
     * Restore a prefixed ontology from a specific backup file. 
     * This can be used to restore an ontology to its previous state after an 
     * invalid insertion has been detected. Note that caller should delete 
     * the backup file if that is desired. 
     * 
     * @param backupFile - to load
     * @param prefix - to associate with loaded ontology 
     * @return OWLOntology - the restored ontology
     * @throws OWLOntologyStorageException
     * @throws OWLException - if problem with ontology
     * @throws FileNotFoundException if backup doesn't exist, or can't be read
     */
    public OWLOntology restore(File backupFile, String prefix) throws OWLOntologyStorageException, FileNotFoundException, OWLException {
        if (!backupFile.exists()) {
            throw new FileNotFoundException(backupFile.getPath());
        }
        // First we have to dump the old ontology if there is one
        OWLOntology oldOnt = Ontology.instance().getOntology(prefix);
        if (oldOnt != null) {
            IRI ontIRI = oldOnt.getOntologyID().getOntologyIRI();
            OWLImportsDeclaration imports = Ontology.instance().getDataFactory().getOWLImportsDeclaration(ontIRI);
            Ontology.instance().getManager().applyChange(new RemoveImport(Ontology.instance().getRootOntology(), imports));
            Ontology.instance().getManager().removeOntology(oldOnt);
        }
        OWLOntology restored = Ontology.instance().loadOntology(backupFile.getPath(), prefix, true);
        //TODO : TBR (this is so we can view results in protege and then revert them at the end of the test
        /*
        if ("config".equals(prefix)) {
            Ontology.instance().saveOntologyByPrefix(Ontology.instance().getConfig().getConfigPrefix());
        }
        */
        return restored;
    }

    /**
     * List all instances in the ontology of a class, given the leaf class name.
     * interesting examples: instance SEL421 instance BadAuthentication instance
     * owl:Thing
     *
     * @deprecated
     * @param className
     * @return
     */
    public String dumpInstances(String className) {
        OWLOntology oriOnt = getRootOntology();
        Set<OWLNamedIndividual> individual = Ontology.instance().getInstances(className);
        StringBuilder output = new StringBuilder();

        for (OWLNamedIndividual ind : individual) {
            String str = ind.toString();
            String iriStr = ind.getIRI().toString();
            logger.info("individual IRI: " + iriStr);
            output.append(getSimpleName(str)).append(NEW_LINE);
            output.append("  (+) Data property:").append(NEW_LINE);
            Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProp = ind.getDataPropertyValues(oriOnt);            
            for (Map.Entry<OWLDataPropertyExpression,Set<OWLLiteral>> entry: dataProp.entrySet())  {
                OWLDataPropertyExpression key = entry.getKey();
                Set<OWLLiteral> values = entry.getValue();
                str = "    - " + key + ":" + values;
                output.append(getSimpleName(str)).append(NEW_LINE);
            }

            output.append(NEW_LINE).append("  (+) Object property:").append(NEW_LINE);
            Map<OWLObjectPropertyExpression, Set<OWLIndividual>> objectProp = ind.getObjectPropertyValues(oriOnt);
            for (Map.Entry<OWLObjectPropertyExpression, Set<OWLIndividual>> entry: objectProp.entrySet()) {
                str = "    - " + entry.getKey() + ":" + entry.getValue();
                output.append(getSimpleName(str)).append(NEW_LINE);
            }
            output.append(NEW_LINE);
        }

        return output.toString();
    }

    /**
     * Derive an OWLDatatype Restriction from a regex pattern
     * @param regex
     * @return 
     */
    public OWLDatatypeRestriction getOWLDatatypeRestrictionFromRegex(String regex) {
        OWLLiteral lit = dataFactory.getOWLLiteral(regex);               
        OWLFacetRestriction facet = dataFactory.getOWLFacetRestriction(OWLFacet.PATTERN, lit);
        OWLDatatype dr = OWL2DatatypeImpl.getDatatype(OWL2Datatype.XSD_STRING);
        return dataFactory.getOWLDatatypeRestriction(dr, facet);
    }
    
    /**
     * Derive an OWLDatatype Restriction from a value and OWLFacet
     * @param value
     * @param facetType
     * @return 
     */
    public OWLDatatypeRestriction getOWLDatatypeRestriction(Object value, OWLFacet facetType) {
        OWLLiteral valLiteral = null;       
        if (value.getClass().equals(Integer.class)) {
             valLiteral = dataFactory.getOWLLiteral((int)value);
        } else if (value.getClass().equals(String.class)) {
             valLiteral = dataFactory.getOWLLiteral((String)value);   
        } else if (value.getClass().equals(Boolean.class)) {
             valLiteral = dataFactory.getOWLLiteral((Boolean)value);   
        }
        OWLFacetRestriction facet = dataFactory.getOWLFacetRestriction(facetType, valLiteral);
        OWLDatatype dr = OWL2DatatypeImpl.getDatatype(OWL2Datatype.XSD_STRING);
        return dataFactory.getOWLDatatypeRestriction(dr, facet);
    }
    
    /**
     * Get a specific type of integer comparison data restriction
     * @param value
     * @param comparitor
     * @return 
     */
    public OWLDatatypeRestriction getOWLIntegerDatatypeRestriction(int value, String comparitor) {
        OWLLiteral oLit = dataFactory.getOWLLiteral(value);
        OWLFacetRestriction fRestriction = null; 
        OWLDatatype oDatatype = OWL2DatatypeImpl.getDatatype(OWL2Datatype.XSD_INTEGER);
        
        switch (comparitor) {
            case "gt":
                fRestriction = dataFactory.getOWLFacetRestriction(OWLFacet.MIN_EXCLUSIVE, oLit);
                break;
            case "lt":
                fRestriction = dataFactory.getOWLFacetRestriction(OWLFacet.MAX_EXCLUSIVE, oLit);
                break;
            case "gte":
                fRestriction = dataFactory.getOWLFacetRestriction(OWLFacet.MIN_INCLUSIVE, oLit);
                break;
            case "lte":
                fRestriction = dataFactory.getOWLFacetRestriction(OWLFacet.MAX_INCLUSIVE, oLit);
                break;
            case "eq":
            case "neq":
                fRestriction = dataFactory.getOWLFacetRestriction(OWLFacet.LENGTH, oLit);
                break;
            default:
                break;
        }
        
        if (fRestriction == null || oDatatype == null) {
            return null;
        } else {
            return dataFactory.getOWLDatatypeRestriction(oDatatype, fRestriction);
        }
    }

    /**
     * Determine if a class exists in the ontology
     * @param name
     * @return 
     */
    boolean classExists(String name) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(true);
        Set<OWLClass> classes = new HashSet<>();       
        for (Object ont : ontArr) {
            classes.addAll(Ontology.instance().getOntology((String)ont).getClassesInSignature());
        }
        
        boolean found = false;
        for (OWLClass c : classes) {
            if (found == false) {
                if (c.toString().endsWith("#" + name + ">")) {
                    found = true;
                }
            }
        }
        
        return found;
    }
    
    /**
     * Find a class based on String value
     * Optionally prioritize the core ontology if there are multiple instances found
     * @param name
     * @param prioritizeRoot
     * @return 
     */
    OWLClass findClass(String name, boolean prioritizeRoot) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(false);
        Set<OWLClass> classes = new HashSet<>();       
        for (Object ont : ontArr) {
            classes.addAll(Ontology.instance().getOntology((String)ont).getClassesInSignature());
        }
        
        boolean found = false;
        OWLClass cls = null;
        if (prioritizeRoot == true) {
            Set<OWLClass> rootClasses = Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getClassesInSignature();
            for (OWLClass c : rootClasses) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        cls = c;
                    }
                }
            }
            for (OWLClass c : classes) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        cls = c;
                    }
                }
            }
        } else {
            classes.addAll(Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getClassesInSignature());
            for (OWLClass c : classes) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        cls = c;
                    }
                }
            }
        }
        
        return cls;
    }

    /**
     * Determine if an individual exists in the ontology
     * @param name
     * @return 
     */
    boolean individualExists(String name) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(true);
        Set<OWLIndividual> individuals = new HashSet<>();       
        for (Object ont : ontArr) {
            OWLOntology oOnt = Ontology.instance().getOntology((String)ont);
            Set<OWLNamedIndividual> inds = oOnt.getIndividualsInSignature();
            individuals.addAll(inds);
        }
        
        boolean found = false;
        for (OWLIndividual c : individuals) {
            if (found == false) {
                if (c.toString().endsWith("#" + name + ">")) {
                    found = true;
                }
            }
        }
        
        return found;
    }
    
    /**
     * Find an individual based on String value
     * Optionally prioritize the core ontology if there are multiple instances found
     * @param name
     * @param prioritizeRoot
     * @return 
     */
    OWLIndividual findIndividual(String name, boolean prioritizeRoot) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(false);
        Set<OWLIndividual> individuals = new HashSet<>();       
        for (Object ont : ontArr) {
            individuals.addAll(Ontology.instance().getOntology((String)ont).getIndividualsInSignature());
        }
        
        boolean found = false;
        OWLIndividual ind = null;
        if (prioritizeRoot == true) {
            Set<OWLNamedIndividual> rootIndividuals = Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getIndividualsInSignature();
            for (OWLIndividual c : rootIndividuals) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        ind = c;
                    }
                }
            }
            for (OWLIndividual c : individuals) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        ind = c;
                    }
                }
            }
        } else {
            individuals.addAll(Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getIndividualsInSignature());
            for (OWLIndividual c : individuals) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        ind = c;
                    }
                }
            }
        }
        
        return ind;
    }

    /**
     * Determine if a data property exists in the ontology
     * @param name
     * @return 
     */
    boolean dataPropertyExists(String name) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(true);
        Set<OWLDataProperty> dps = new HashSet<>();       
        for (Object ont : ontArr) {
            dps.addAll(Ontology.instance().getOntology((String)ont).getDataPropertiesInSignature());
        }
        
        boolean found = false;
        for (OWLDataProperty c : dps) {
            if (found == false) {
                if (c.toString().endsWith("#" + name + ">")) {
                    found = true;
                }
            }
        }
        
        return found;
    }
    
    /**
     * Find a data property based on String value
     * Optionally prioritize the core ontology if there are multiple instances found
     * @param name
     * @param prioritizeRoot
     * @return 
     */
    OWLDataProperty findDataProperty(String name, boolean prioritizeRoot) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(false);
        Set<OWLDataProperty> odps = new HashSet<>();       
        for (Object ont : ontArr) {
            odps.addAll(Ontology.instance().getOntology((String)ont).getDataPropertiesInSignature());
        }
        Set<OWLDataProperty> rootdps = Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getDataPropertiesInSignature();
        
        boolean found = false;
        OWLDataProperty dp = null;
        if (prioritizeRoot == true) {         
            for (OWLDataProperty c : rootdps) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        dp = c;
                    }
                }
            }
            for (OWLDataProperty c : odps) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        dp = c;
                    }
                }
            }
        } else {
            odps.addAll(Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getDataPropertiesInSignature());
            for (OWLDataProperty c : odps) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        dp = c;
                    }
                }
            }
        }
        
        return dp;
    }
    
    /**
     * Determine if a object property exists in the ontology
     * @param name
     * @return 
     */
    boolean objectPropertyExists(String name) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(true);
        Set<OWLObjectProperty> ops = new HashSet<>();       
        for (Object ont : ontArr) {
            ops.addAll(Ontology.instance().getOntology((String)ont).getObjectPropertiesInSignature());
        }
        
        boolean found = false;
        for (OWLObjectProperty c : ops) {
            if (found == false) {
                if (c.toString().endsWith("#" + name + ">")) {
                    found = true;
                }
            }
        }
        
        return found;
    }
    
    /**
     * Find an object property based on String value
     * Optionally prioritize the core ontology if there are multiple instances found
     * @param name
     * @param prioritizeRoot
     * @return 
     */
    OWLObjectProperty findObjectProperty(String name, boolean prioritizeRoot) {
        ArrayList ontArr = config.getAdditionalOntologyPrefixes(false);
        Set<OWLObjectProperty> oops = new HashSet<>();       
        for (Object ont : ontArr) {
            oops.addAll(Ontology.instance().getOntology((String)ont).getObjectPropertiesInSignature());
        }
        Set<OWLObjectProperty> rootops = Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getObjectPropertiesInSignature();
        
        boolean found = false;
        OWLObjectProperty dp = null;
        if (prioritizeRoot == true) {         
            for (OWLObjectProperty c : rootops) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        dp = c;
                    }
                }
            }
            for (OWLObjectProperty c : oops) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        dp = c;
                    }
                }
            }
        } else {
            oops.addAll(Ontology.instance().getOntology(Ontology.instance().getConfig().getCorePrefix()).getObjectPropertiesInSignature());
            for (OWLObjectProperty c : oops) {
                if (found == false) {
                    if (c.toString().endsWith("#" + name + ">")) {
                        found = true;
                        dp = c;
                    }
                }
            }
        }
        
        return dp;
    }
    
    /**
     * This has a pretty specific use so far, which is just to test if an object property has range 'Status'
     * the Range will generally be a class
     * @param property
     * @param range
     * @return 
     */
    public boolean objectPropertyHasRange(String property, String range) {
        ArrayList ontologyPrefixes = config.getAdditionalOntologyPrefixes(false);
        Set<OWLOntology> ontologyArray = new HashSet<>();    
        ontologyArray.add(Ontology.instance().getRootOntology());
        for (Object prefix : ontologyPrefixes) {
            ontologyArray.add(Ontology.instance().getOntology((String) prefix));
        }
        
        OWLObjectProperty op = Ontology.instance().findObjectProperty(property, true);
        OWLClass rg = Ontology.instance().findClass(range, true);      
        Set<OWLClassExpression> propExpressions = op.getRanges(ontologyArray);
        
        String rgStr = rg.toString();
        for (OWLClassExpression ex : propExpressions) {
            String exStr = ex.toString();
            
            if (rgStr.equals(exStr) || rgStr.endsWith(exStr.split("#")[1])) {
                return true;
            }
        }
        
        return false;
    }
    
    /**
     * Given an individual name, remove any associated content in a selected ontology
     * This includes object properties, data properties, and types
     * @param prefix 
     * @param indName 
     */
    public void clearIndividual(String prefix, String indName) {
        OWLOntology ont = this.getOntology(prefix);
        
        if (ont == null) {
            logger.error("Unable to clear individual, invalid ontology");
            return;
        }
         
        Individual ind = Ontology.instance().getIndividual(indName, ont);
        OWLEntityRemover remover = new OWLEntityRemover(this.manager, Collections.singleton(ont));
        remover.visit(ind.getOWLIndividual());
        manager.applyChanges(remover.getChanges());
    }
    
    /**
     * This function replaces an ontology based on its prefix
     * This is used as an easier way of clearing ontologies back to their original
     * state before adding new policy or configuration since it is replaced entirely
     * on every request.
     * @param prefix 
     */
    public void replaceOntology(String prefix) {
        if (prefix.equals(this.config.getPolicyPrefix())) {
            //Policy requires a few test items and a few default classes to work properly.
            List<String> policyToSave = new ArrayList();

            //Test individual not to remove
            policyToSave.add("test421a");
            policyToSave.add("test421b");
            policyToSave.add("test421c");
            
            //Classes required to mkae poilcy CWR's work
            policyToSave.add("ClosedWorldReasoner");
            policyToSave.add("TESTClosedWorldReasoner");

            //If it's policy, we need to make sure we create a few classes first
            OWLOntology ont = this.getOntology(prefix);
            
            if (ont == null) {
                logger.info("Ontology was null, cant replace");
                return;
            }
            
            OWLEntityRemover remover = new OWLEntityRemover(this.manager, Collections.singleton(ont));
            
            for (OWLDataProperty dp : ont.getDataPropertiesInSignature()) {
                if (!policyToSave.contains(dp.toString().split("#")[1].replace(">", ""))) {
                    dp.accept(remover);
                }
            } 
            for (OWLObjectProperty op : ont.getObjectPropertiesInSignature()) {
                if (!policyToSave.contains(op.toString().split("#")[1].replace(">", ""))) {
                    op.accept(remover);
                }
            }
            for (OWLNamedIndividual ind : ont.getIndividualsInSignature()) {
                if (!policyToSave.contains(ind.toString().split("#")[1].replace(">", ""))) {
                    ind.accept(remover);
                }
            }
            for (OWLClass cls : ont.getClassesInSignature()) {
                if (!policyToSave.contains(cls.toString().split("#")[1].replace(">", ""))) {
                    cls.accept(remover);
                }
            }
            
            manager.applyChanges(remover.getChanges());
        } else {
            OWLOntology ont = this.getOntology(prefix);
            
            if (ont == null) {
                logger.info("Ontology was null, cant replace");
                return;
            }
            
            OWLEntityRemover remover = new OWLEntityRemover(this.manager, Collections.singleton(ont));
            
            for (OWLDataProperty dp : ont.getDataPropertiesInSignature()) {
                dp.accept(remover);
            } 
            for (OWLObjectProperty op : ont.getObjectPropertiesInSignature()) {
                op.accept(remover);
            }
            for (OWLNamedIndividual ind : ont.getIndividualsInSignature()) {
                ind.accept(remover);
            }
            for (OWLClass cls : ont.getClassesInSignature()) {
                cls.accept(remover);
            }
            
            manager.applyChanges(remover.getChanges());
        }
    }
}
