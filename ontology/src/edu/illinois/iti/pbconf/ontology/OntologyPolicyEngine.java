/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import java.util.ArrayList;
import java.util.HashSet;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLDataComplementOf;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataUnionOf;
import org.semanticweb.owlapi.model.OWLDatatypeRestriction;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.vocab.OWLFacet;

/**
 * Ontology policy engine - creates ontology statements based on policy set from the parser.
 * @author josephdigiovanna
 */
public final class OntologyPolicyEngine {   
    //Handles whether or not we need to add closed world reasoners under a specific test prefix.
    //Strictly for unit tests for closed world reasoners.
    boolean inTestMode = false;
    
    /**
     * Default constructor for the policy engine, which processes policy to create ontology statements
     */
    public OntologyPolicyEngine() {}
  
     /**
     * Get an array of datatype restrictions for use in axioms
     *
     * @param propObj
     * @return
     */
    public ArrayList<OWLDatatypeRestriction> getOWLDataPropertyDatatypeRestrictions(JSONObject propObj) {
        ArrayList<OWLDatatypeRestriction> dtrs = new ArrayList<>();
          //Now get the object we're operating on
        //From there, derive the Owl datatype restriction
        switch (propObj.getString("predicate")) {
            case "complexity":
                switch ((String) propObj.get("value")) {
                    case "LOWERCASE":
                        dtrs.add(Ontology.instance().getOWLDatatypeRestrictionFromRegex("[a-z]*"));
                        break;
                    case "UPPERCASE":
                        dtrs.add(Ontology.instance().getOWLDatatypeRestrictionFromRegex("[A-Z]*"));
                        break;
                    case "MIXEDCASE":
                        dtrs.add(Ontology.instance().getOWLDatatypeRestrictionFromRegex("[a-z]*"));
                        dtrs.add(Ontology.instance().getOWLDatatypeRestrictionFromRegex("[A-Z]*"));
                        break;
                    default:
                        break;
                }
                break;
            case "min-length":
                dtrs.add(Ontology.instance().getOWLDatatypeRestriction(propObj.get("value"), OWLFacet.MIN_LENGTH));
                break;
            case "max-length":
                dtrs.add(Ontology.instance().getOWLDatatypeRestriction(propObj.get("value"), OWLFacet.MAX_LENGTH));
                break;
            case "gt":
            case "lt":
            case "gte":
            case "lte":
            case "eq":
            case "neq":    
                dtrs.add(Ontology.instance().getOWLIntegerDatatypeRestriction(propObj.getInt("value"), propObj.getString("predicate")));
                break;
            default:
                break;
        }

        return dtrs;
    }

    /**
     * Get an OWLAxiom from an array of data property restrictions
     *
     * @param propObj
     * @param dataProp
     * @param dtrs
     * @return
     */
    public OWLAxiom getOWLAxiomFromDataRestrictions(JSONObject propObj, OWLDataProperty dataProp, ArrayList<OWLDatatypeRestriction> dtrs) {
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLAxiom mAxiom = null;

        if ("string".equals(propObj.getString("valueType"))) {
            if ("MIXEDCASE".equals(propObj.getString("value"))) {
                OWLDataUnionOf dtrsUnion = odf.getOWLDataUnionOf(new HashSet<>(dtrs));
                OWLDataComplementOf dco = odf.getOWLDataComplementOf(dtrsUnion);
                mAxiom = odf.getOWLDataPropertyRangeAxiom(dataProp, dco);
            } else {
                if (dtrs.isEmpty()) {
                    mAxiom = null;
                } else if (dtrs.size() == 1) {
                    mAxiom = odf.getOWLDataPropertyRangeAxiom(dataProp, dtrs.get(0));
                }
            }
        } else {
            if (dtrs.isEmpty()) {
                mAxiom = null;
            } else if (dtrs.size() == 1) {
                if (propObj.getString("predicate").equals("neq")) {
                    OWLDataComplementOf dc = odf.getOWLDataComplementOf(dtrs.get(0));
                    mAxiom = odf.getOWLDataPropertyRangeAxiom(dataProp, dc);
                } else {
                    mAxiom = odf.getOWLDataPropertyRangeAxiom(dataProp, dtrs.get(0));
                }
            }
        }

        return mAxiom;
    }

    /**
     * Currently, we support classes having or not having a specific property
     * Next will be support of individuals
     *
     * @param ont
     * @param propObj
     * @return
     */
    public String addCWRAxiom(OWLOntology ont, JSONObject propObj) {
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        boolean cwrValid = false;
        boolean cwrSaved = false;

        ClosedWorld.Validator val = ClosedWorld.getValidator(propObj.getString("cwrType"));
        if (inTestMode == true) {
            val.overrideClass("TESTClosedWorldReasoner", "tcwr");
        }

        if (propObj.getString("cwrType").contains("MustHaveProperty") || propObj.getString("cwrType").contains("MustNotHaveProperty")) {
            if (propObj.getString("cwrType").equals("IndividualMustHaveProperty") || propObj.getString("cwrType").equals("IndividualMustNotHaveProperty")) {
                OWLIndividual ind = odf.getOWLNamedIndividual(Ontology.instance().getIRI(propObj.getString("translatedSubject")));
                OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI(propObj.getString("translatedObject")));
                cwrValid = val.validateRequest(ind, odp, "add");
                if (cwrValid == true) {
                    String nextName = ClosedWorld.getNextClosedWorldReasonerName(val.getCurrentClass(), val.getCurrentPrefix());
                    cwrSaved = val.saveLastRequest(nextName);
                }
            } else {
                OWLClass cls = odf.getOWLClass(Ontology.instance().getIRI(propObj.getString("translatedSubject")));
                OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI(propObj.getString("translatedObject")));
                cwrValid = val.validateRequest(cls, odp, "add");
                if (cwrValid == true) {
                    String nextName = ClosedWorld.getNextClosedWorldReasonerName(val.getCurrentClass(), val.getCurrentPrefix());
                    cwrSaved = val.saveLastRequest(nextName);
                }
            }
        } else if (propObj.getString("cwrType").contains("PropertyRestrictions")) {
            cwrValid = val.validateJSONObject(propObj, "add");
            if (cwrValid == true) {
                String nextName = ClosedWorld.getNextClosedWorldReasonerName(val.getCurrentClass(), val.getCurrentPrefix());
                cwrSaved = val.saveLastRequest(nextName);
            }
        }

        String cls = "ClosedWorldReasoner";
        String prefix = "cwr";
        
        if (inTestMode) {
            cls = "TESTClosedWorldReasoner";
            prefix = "tcwr";
        }
        
        String expl = "";
        if (cwrValid && cwrSaved) {
            return "VALID";
        } else {
            return "INVALID:" + val.getExplanation();
        }
    }

    /**
     * Add a standard Policy element using the OWL API
     *
     * @param ont
     * @param propObj
     * @return
     */
    public String addStandardPolicy(OWLOntology ont, JSONObject propObj) {
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLOntologyManager mgr = ont.getOWLOntologyManager();
        ArrayList<OWLDatatypeRestriction> dataRestrictions = null;
        OWLAxiom mAxiom = null;

        String subj = propObj.getString("translatedSubject");    
        OWLDataProperty dataProp = odf.getOWLDataProperty(Ontology.instance().getIRI(subj));
        dataRestrictions = this.getOWLDataPropertyDatatypeRestrictions(propObj);
        mAxiom = this.getOWLAxiomFromDataRestrictions(propObj, dataProp, dataRestrictions);

        if (mAxiom != null) {
            mgr.addAxiom(ont, mAxiom);
            //Since we're adding standard policy, only need to check against standard consistency
            if (Ontology.instance().isConstistent()) {
                return "VALID";
            } else {
                return "INVALID:" + Ontology.instance().getFriendlyExplanation(Ontology.instance().getExplanation());
            }
        } else {
            String translatedPredicate = propObj.getString("predicate");
            String translatedObject = propObj.getString("translatedObject");
            return "INVALID:Unable to construct OWL Axiom from (s, p, o) : " + subj + ", " + translatedPredicate + ", " + translatedObject; 
        }
    }
    
    /**
     * For a given ontology, and target (class, individual, object property, data property), use the subject,
     * predicate, object triple to generate and store an axiom in the appropriate location
     * 
     * @param ont
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @param testMode
     */
    public void addAxiom(OWLOntology ont, String target, String subject, String predicate, Object object, boolean testMode) {
        //Update on a per event basis if we're in test mode
        inTestMode = testMode;

        if (ont == null) {
            return;
        }

        PBConfParser pbParser = new PBConfParser();
        JSONObject propObj = pbParser.getAxiomParts(target, subject, predicate, object);
        String result = "";
        
        if (null != propObj.getString("policyType")) {
            switch (propObj.getString("policyType")) {
                case "data":
                case "object":
                case "individual":
                case "class":    
                    result = addStandardPolicy(ont, propObj);
                    break;
                case "cwr":
                    result = addCWRAxiom(ont, propObj);
                    break;
                case "invalid":
                    result = "INVALID:Could not create policy from statement";
                    break; 
                default:
                    //Return false, we don't know how to process the request
                    result = "INVALID:Could not create policy from statement";
                    break;        
            }
        } else {
            result = "INVALID:Unable to parse policy components";
        }   
    }
}
