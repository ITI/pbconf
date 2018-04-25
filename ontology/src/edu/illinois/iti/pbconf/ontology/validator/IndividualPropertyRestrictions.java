/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology.validator;

import edu.illinois.iti.pbconf.ontology.ClosedWorld;
import edu.illinois.iti.pbconf.ontology.Individual;
import edu.illinois.iti.pbconf.ontology.Ontology;
import edu.illinois.iti.pbconf.ontology.OntologyConfig;
import java.io.File;
import java.io.FileNotFoundException;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.Set;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLException;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLObjectPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyStorageException;

/**
 *
 * @author anderson
 */
public class IndividualPropertyRestrictions implements ClosedWorld.Validator {
    static final Logger logger = Logger.getLogger(IndividualPropertyRestrictions.class.getName().replaceFirst(".+\\.", ""));
    boolean lastRequestValid = false;
    Object[] lastRequestArgs;
    JSONObject lastRequestJSON;
    
    String explanation = "";
    //This is overridable and is important because we can set it up for testing
    String validatorClass = "ClosedWorldReasoner";
    String validatorIndPrefix = "cwr";
    
    /**
     * Force a different class / prefix, usually prepending TEST/t
     * @param cls
     * @param prefix 
     */
    @Override
    public void overrideClass(String cls, String prefix) {
        validatorClass = cls;
        validatorIndPrefix = prefix;
    }
    
    /**
     * 
     * @return 
     */
    @Override
    public String getCurrentClass() {
        return validatorClass;
    }
    
    /**
     * 
     * @return 
     */
    @Override
    public String getCurrentPrefix() {
        return validatorIndPrefix;
    }
    
    /**
     * Validate a formatted JSON axiom object
     * @param propObj
     * @param type
     * @return 
     */
    @Override
    public boolean validateJSONObject(JSONObject propObj, String type) {
        ArrayList invalidInd = new ArrayList();
        boolean valid = true;       
        String individualStr = propObj.getString("translatedTarget");
        ArrayList ontArr = Ontology.instance().getConfig().getAdditionalOntologyPrefixes(false);
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        
        OWLIndividual i = odf.getOWLNamedIndividual(Ontology.instance().getIRI(individualStr));
        
        //We need to test the individual for the properties mentioned in the json object                   
        Set<OWLClassExpression> types = new HashSet<>();
        Set<OWLObjectPropertyAssertionAxiom> oAssertions = new HashSet<>();
        Set<OWLDataPropertyAssertionAxiom> dAssertions = new HashSet<>();

        types.addAll(i.getTypes(Ontology.instance().getRootOntology()));
        oAssertions.addAll(Ontology.instance().getRootOntology().getObjectPropertyAssertionAxioms(i));
        dAssertions.addAll(Ontology.instance().getRootOntology().getDataPropertyAssertionAxioms(i));

        for (Object ont : ontArr) {
            String prefix = (String) ont;
            types.addAll(i.getTypes(Ontology.instance().getOntology(prefix)));
            oAssertions.addAll(Ontology.instance().getOntology(prefix).getObjectPropertyAssertionAxioms(i));
            dAssertions.addAll(Ontology.instance().getOntology(prefix).getDataPropertyAssertionAxioms(i));
        }

        boolean subValid = true;

        //Test matching data properties
        for (OWLDataPropertyAssertionAxiom ax: dAssertions) {
            if (ClosedWorld.testDataProperty(propObj, ax) == false) {
                subValid = false;
            }
        }

        //Test matching object properties
        for (OWLObjectPropertyAssertionAxiom ax: oAssertions) {
            if (ClosedWorld.testObjectProperty(propObj, ax) == false) {
                subValid = false;
            }
        }

        //Test for matching type
        for (OWLClassExpression ax : types) {
            if (ClosedWorld.testType(propObj, ax) == false) {
                subValid = false;
            }
        }

        valid = valid && subValid;
        if (!subValid) {
            String indStr = i.toString();
            indStr = indStr.replace("<", "\"");
            indStr = indStr.replace(">", "\"");
            //String propStr = "\"" + opd.toString().split("#")[1] + "\"";
            //String obj = "{\"individual\":" + indStr + ",\"property\":" + propStr + "}";                     
            String obj = "{\"individual\":" + indStr + "}";
            invalidInd.add(obj);
        }
        
        if (!valid) {
            explanation = invalidInd.toString();
            lastRequestValid = false;
        } else {
            explanation = "";
            lastRequestValid = true;
            lastRequestJSON = propObj;
        }
        
        if (type.equals("add")) {
            explanation = "";
            lastRequestValid = true;
            lastRequestJSON = propObj;
        }

        return valid;
    }
    
    /**
     * In this case, Object args will be OWLClass, OWLDataProperty
     * @param args
     * @return 
     */
    @Override
    public boolean validateRequest(Object... args) {
        if (args.length > 1) {
            OWLIndividual ind = (OWLIndividual) args[0];
            OWLDataProperty opd = (OWLDataProperty) args[1];
            ArrayList invalidInd = new ArrayList();
            boolean valid = true;         
            
            Set<OWLDataPropertyAssertionAxiom> properties = Ontology.instance().getRootOntology().getDataPropertyAssertionAxioms(ind);
            Set<OWLDataPropertyAssertionAxiom> configProperties = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()).getDataPropertyAssertionAxioms(ind);
            Set<OWLDataPropertyAssertionAxiom> policyProperties = Ontology.instance().getOntology(Ontology.instance().getConfig().getPolicyPrefix()).getDataPropertyAssertionAxioms(ind);

            boolean subValid = false;

            for (OWLDataPropertyAssertionAxiom ax: properties) {
                String prop = ax.getProperty().toString();
                String odpStr = opd.toString();
                if (prop.equals(odpStr)) {
                    subValid = true;
                }
            }

            for (OWLDataPropertyAssertionAxiom ax: configProperties) {
                String prop = ax.getProperty().toString();
                String odpStr = opd.toString();
                if (prop.equals(odpStr) || prop.endsWith(odpStr.split("#")[1])) {
                    subValid = true;
                }
            }

            for (OWLDataPropertyAssertionAxiom ax: policyProperties) {
                String prop = ax.getProperty().toString();
                String odpStr = opd.toString();
                if (prop.equals(odpStr) || prop.endsWith(odpStr.split("#")[1])) {
                    subValid = true;
                }
            }

            valid = valid && subValid;
            if (!subValid) {
                String indStr = ind.toString();
                indStr = indStr.replace("<", "\"");
                indStr = indStr.replace(">", "\"");
                String propStr = "\"" + opd.toString().split("#")[1] + "\"";
                String obj = "{\"individual\":" + indStr + ",\"property\":" + propStr + "}";                     
                invalidInd.add(obj);
            }
            
            if (!valid) {
                explanation = invalidInd.toString();
                lastRequestValid = false;
            } else {
                explanation = "";
                lastRequestValid = true;
                lastRequestArgs = args;
            }
 
            return valid;
        } else {
            explanation = "Invalid arguement(s)";
            return false;
        }
    }
    
    /**
     * Get an explanation if one exists
     * @return
     */
    @Override
    public String getExplanation() {
        if ("".equals(explanation)) {
            return "";
        }
        //At this point, we know we have a failure for one or more individuals
        //We will report the validator name, and an array of individuals that failed
        //This will be done in stringified JSON for processing on the PBConf side.
        String fullExplanation = "{\"failedValidator\":\"";
        fullExplanation += ClassPropertyRestrictions.class.getName().replaceFirst(".+\\.", "");
        
        fullExplanation += "\",\"failures\":";
        fullExplanation += explanation;
        
        fullExplanation += "}";
        return fullExplanation;      
    }
    
    /**
     * If the last request was valid, optionally save it to the ontology
     * Allowing this to be optional for cases where we just what to validate against
     * something, not necessarily make it permanent 
     * @param nextCWR
     * @return 
     */
    @Override
    public boolean saveLastRequest(String nextCWR) {
        if (lastRequestValid == false || (lastRequestArgs == null && lastRequestJSON.length() == 0) || (lastRequestArgs != null && lastRequestArgs.length == 0 && lastRequestJSON.length() == 0)) {
            return false;
        }
        
        OntologyConfig cfg = Ontology.instance().getConfig(); 
        //Backup before attempting to alter the ontology
        File cwrBackup;
        try {
            cwrBackup = Ontology.instance().backup(cfg.get("closedWorldOntology"));
        } catch (OWLOntologyStorageException | FileNotFoundException ex) {
            logger.error("Unable to back up ontology : " + ex.toString());
            return false;
        }
        
        try {
            String cwPrefix = cfg.getClosedWorldPrefixStr();
            OWLOntology targetOnt = Ontology.instance().getOntology(cfg.get("closedWorldOntology"));

            IRI cwrIRI = IRI.create(cwPrefix + "#" + validatorClass);
            Individual ind = Ontology.instance().getIndividual(nextCWR, targetOnt);        
            ind.setClass(cwrIRI);

            IRI reasonerNameIRI = IRI.create(cwPrefix + "#" + "hasReasonerName");
            IRI individualTargetIRI = IRI.create(cwPrefix + "#" + "hasIndividualTarget");
            IRI subjectIRI = IRI.create(cwPrefix + "#" + "subject");
            IRI predicateIRI = IRI.create(cwPrefix + "#" + "predicate");
            IRI objectIRI = IRI.create(cwPrefix + "#" + "object");
            
            String individualName = lastRequestJSON.getString("translatedTarget");
            String subject = lastRequestJSON.getString("translatedSubject");
            String predicate = lastRequestJSON.getString("predicate");
            String object = lastRequestJSON.getString("originalObject");
            
            ind.setProperty(reasonerNameIRI, "IndividualPropertyRestrictions");
            ind.setProperty(individualTargetIRI, individualName);
            ind.setProperty(subjectIRI, subject);
            ind.setProperty(predicateIRI, predicate);
            ind.setProperty(objectIRI, object);

            //isConsistent factors in validators now, so this should be fine
            if (Ontology.instance().isConstistent() == false) {
                explanation = Ontology.instance().getFriendlyExplanation(Ontology.instance().getExplanation());
                explanation = explanation.replace("\n","|");
                explanation = JSONObject.quote(explanation);

                // Now restore original config
                Ontology.instance().restore(cwrBackup, cfg.get("closedWorldOntology"));
            }
            if (cwrBackup != null && !cwrBackup.delete()) {
                logger.error("Could not delete backup file:"+cwrBackup);
            }
            
            return true;
        } catch (OWLException | FileNotFoundException ex) {
            logger.error(ex);
            explanation = "Ontology exception: " + ex.toString();
            return false;
        }
    }
}
