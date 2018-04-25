/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import static edu.illinois.iti.pbconf.ontology.ClosedWorldTest.logger;
import java.io.FileNotFoundException;
import java.util.ArrayList;
import java.util.logging.Level;
import org.apache.log4j.BasicConfigurator;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.junit.AfterClass;
import org.junit.Before;
import org.junit.BeforeClass;
import org.semanticweb.owlapi.model.OWLException;
import static org.junit.Assert.*;
import org.junit.Test;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDatatypeRestriction;

/**
 *
 * @author anderson
 */
public class OntologizerTest {
    static Logger logger = Logger.getLogger(OntologizerTest.class.getName().replaceFirst(".+\\.",""));
    
    /**
     * Setup configuration
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
     * Load configuration file and ontologies
     */
    @Before
    public void setUp() {
        Ontology.instance().reset();
//        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig("pbconf.json");
        String c = System.getProperty("user.dir");
        c  = c.concat("/config/pbconf.json");
        OntologyConfig.JSONConfig cfg = new OntologyConfig.JSONConfig(c);
        
        try {
            Ontology.instance().initialize(cfg);
        } catch (FileNotFoundException ex) {
            java.util.logging.Logger.getLogger(OntologyTest.class.getName()).log(Level.SEVERE, null, ex);
            logger.error("base ontology file", ex);
        } catch (OWLException ex) {
            logger.error("OWLException", ex);
        }
    }

    /**
     * Make sure we can get the appropriate ontologizer
     * @throws Exception
     */
    @Test
    public void test_getOntologizer() throws Exception {
        Ontologizer gizer = Ontologizer.getOntologizer("sel421");
        logger.info(gizer.getClass().toString());
        String clsName = gizer.getClass().toString();
        assertTrue(clsName.contains("Ontologizer$SEL421Ontologizer"));
    }
    
    /**
     * Get and validate an array of data property restrictions based on input
     * @throws Exception 
     */
    @Test
    public void test_getOWLDataPropertyDatatypeRestrictions_0() throws Exception {
        ArrayList<OWLDatatypeRestriction> dtrs = null;        
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "min-length";
        Object o = "6";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        PBConfParser pbParser = new PBConfParser();
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String output = dtrs.get(0).toString();
        assertTrue("DataRangeRestriction(xsd:string facetRestriction(minLength \"6\"^^xsd:integer))".equals(output));
    }
    
    /**
     * Get and validate an array of data property restrictions based on input
     * @throws Exception 
     */
    @Test
    public void test_getOWLDataPropertyDatatypeRestrictions_1() throws Exception {
        ArrayList<OWLDatatypeRestriction> dtrs = null;    
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "max-length";
        Object o = "12";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        PBConfParser pbParser = new PBConfParser();
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String output = dtrs.get(0).toString();
        assertTrue("DataRangeRestriction(xsd:string facetRestriction(maxLength \"12\"^^xsd:integer))".equals(output));
    }
    
    /**
     * Get and validate an array of data property restrictions based on input
     * @throws Exception 
     */
    @Test
    public void test_getOWLDataPropertyDatatypeRestrictions_2() throws Exception {
        ArrayList<OWLDatatypeRestriction> dtrs = null;     
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "complexity";
        Object o = "LOWERCASE";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        PBConfParser pbParser = new PBConfParser();
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String output = dtrs.get(0).toString();
        assertTrue("DataRangeRestriction(xsd:string facetRestriction(pattern \"[a-z]*\"^^xsd:string))".equals(output));
    }
    
    /**
     * Get and validate an array of data property restrictions based on input
     * @throws Exception 
     */
    @Test
    public void test_getOWLDataPropertyDatatypeRestrictions_3() throws Exception {
        ArrayList<OWLDatatypeRestriction> dtrs = null;  
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "complexity";
        Object o = "UPPERCASE";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        PBConfParser pbParser = new PBConfParser();
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String output = dtrs.get(0).toString();
        assertTrue("DataRangeRestriction(xsd:string facetRestriction(pattern \"[A-Z]*\"^^xsd:string))".equals(output));
    }
    
    /**
     * Get and validate an array of data property restrictions based on input
     * @throws Exception 
     */
    @Test
    public void test_getOWLDataPropertyDatatypeRestrictions_4() throws Exception {
        ArrayList<OWLDatatypeRestriction> dtrs = null;  
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "complexity";
        Object o = "MIXEDCASE";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        PBConfParser pbParser = new PBConfParser();
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String outputA = dtrs.get(0).toString();
        String outputB = dtrs.get(1).toString();
        assertTrue("DataRangeRestriction(xsd:string facetRestriction(pattern \"[a-z]*\"^^xsd:string))".equals(outputA));
        assertTrue("DataRangeRestriction(xsd:string facetRestriction(pattern \"[A-Z]*\"^^xsd:string))".equals(outputB));
    }
    
    /**
     * Make sure we can generate an axiom from restrictions
     * @throws Exception 
     */
    @Test
    public void test_getOWLAxiomFromDataRestrictions_0() throws Exception {
        OWLDataFactory odf = Ontology.instance().getDataFactory(); 
        PBConfParser pbParser = new PBConfParser();         
        ArrayList<OWLDatatypeRestriction> dtrs = null;
        OWLAxiom mAxiom = null;   
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "min-length";
        Object o = "6";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String tProp = propObj.getString("translatedSubject");
        IRI propIRI = Ontology.instance().getIRI(tProp);
        OWLDataProperty dataProp = odf.getOWLDataProperty(propIRI);
        
        mAxiom = ope.getOWLAxiomFromDataRestrictions(propObj, dataProp, dtrs);
        String axiomStr = mAxiom.toString();
        assertTrue("DataPropertyRange(<http://iti.illinois.edu/iti/pbconf/core#hasLvl1Pwd> DataRangeRestriction(xsd:string facetRestriction(minLength \"6\"^^xsd:integer)))".equals(axiomStr));
    }
    
    /**
     * Make sure we can generate an axiom from restrictions
     * @throws Exception 
     */
    @Test
    public void test_getOWLAxiomFromDataRestrictions_1() throws Exception {
        OWLDataFactory odf = Ontology.instance().getDataFactory(); 
        PBConfParser pbParser = new PBConfParser();         
        ArrayList<OWLDatatypeRestriction> dtrs = null;
        OWLAxiom mAxiom = null;   
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "max-length";
        String o = "12";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String tProp = propObj.getString("translatedSubject");
        IRI propIRI = Ontology.instance().getIRI(tProp);
        OWLDataProperty dataProp = odf.getOWLDataProperty(propIRI);
        
        mAxiom = ope.getOWLAxiomFromDataRestrictions(propObj, dataProp, dtrs);
        String axiomStr = mAxiom.toString();
        assertTrue("DataPropertyRange(<http://iti.illinois.edu/iti/pbconf/core#hasLvl1Pwd> DataRangeRestriction(xsd:string facetRestriction(maxLength \"12\"^^xsd:integer)))".equals(axiomStr));
    }
    
    /**
     * Make sure we can generate an axiom from restrictions
     * @throws Exception 
     */
    @Test
    public void test_getOWLAxiomFromDataRestrictions_2() throws Exception {
        OWLDataFactory odf = Ontology.instance().getDataFactory(); 
        PBConfParser pbParser = new PBConfParser();         
        ArrayList<OWLDatatypeRestriction> dtrs = null;
        OWLAxiom mAxiom = null;   
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "complexity";
        Object o = "LOWERCASE";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String tProp = propObj.getString("translatedSubject");
        IRI propIRI = Ontology.instance().getIRI(tProp);
        OWLDataProperty dataProp = odf.getOWLDataProperty(propIRI);
        
        mAxiom = ope.getOWLAxiomFromDataRestrictions(propObj, dataProp, dtrs);
        String axiomStr = mAxiom.toString();
        assertTrue("DataPropertyRange(<http://iti.illinois.edu/iti/pbconf/core#hasLvl1Pwd> DataRangeRestriction(xsd:string facetRestriction(pattern \"[a-z]*\"^^xsd:string)))".equals(axiomStr));
    }
    
    /**
     * Make sure we can generate an axiom from restrictions
     * @throws Exception 
     */
    @Test
    public void test_getOWLAxiomFromDataRestrictions_3() throws Exception {
        OWLDataFactory odf = Ontology.instance().getDataFactory(); 
        PBConfParser pbParser = new PBConfParser();         
        ArrayList<OWLDatatypeRestriction> dtrs = null;
        OWLAxiom mAxiom = null;   
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "complexity";
        Object o = "UPPERCASE";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String tProp = propObj.getString("translatedSubject");
        IRI propIRI = Ontology.instance().getIRI(tProp);
        OWLDataProperty dataProp = odf.getOWLDataProperty(propIRI);
        
        mAxiom = ope.getOWLAxiomFromDataRestrictions(propObj, dataProp, dtrs);
        String axiomStr = mAxiom.toString();
        assertTrue("DataPropertyRange(<http://iti.illinois.edu/iti/pbconf/core#hasLvl1Pwd> DataRangeRestriction(xsd:string facetRestriction(pattern \"[A-Z]*\"^^xsd:string)))".equals(axiomStr));
    }
    
    /**
     * Make sure we can generate an axiom from restrictions
     * @throws Exception 
     */
    @Test
    public void test_getOWLAxiomFromDataRestrictions_4() throws Exception {
        OWLDataFactory odf = Ontology.instance().getDataFactory(); 
        PBConfParser pbParser = new PBConfParser();         
        ArrayList<OWLDatatypeRestriction> dtrs = null;
        OWLAxiom mAxiom = null;   
        
        String t = "SEL421";
        String s = "password.level1";
        String p = "complexity";
        Object o = "MIXEDCASE";
        
        OntologyPolicyEngine ope = new OntologyPolicyEngine();
        
        JSONObject propObj = pbParser.getAxiomParts(t, s, p, o);
        dtrs = ope.getOWLDataPropertyDatatypeRestrictions(propObj); 
        
        String tProp = propObj.getString("translatedSubject");
        IRI propIRI = Ontology.instance().getIRI(tProp);
        OWLDataProperty dataProp = odf.getOWLDataProperty(propIRI);
        
        mAxiom = ope.getOWLAxiomFromDataRestrictions(propObj, dataProp, dtrs);
        String axiomStr = mAxiom.toString();
        assertTrue("DataPropertyRange(<http://iti.illinois.edu/iti/pbconf/core#hasLvl1Pwd> DataComplementOf(DataUnionOf(DataRangeRestriction(xsd:string facetRestriction(pattern \"[A-Z]*\"^^xsd:string)) DataRangeRestriction(xsd:string facetRestriction(pattern \"[a-z]*\"^^xsd:string)) )))".equals(axiomStr));
    }
}
