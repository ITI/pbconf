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
import java.util.Collections;
import java.util.Set;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLException;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyStorageException;
import org.semanticweb.owlapi.reasoner.NodeSet;
import org.semanticweb.owlapi.reasoner.OWLReasoner;

/**
 *
 * @author anderson
 */
public class ClassMustHaveProperty implements ClosedWorld.Validator {
    static final Logger logger = Logger.getLogger(ClassMustHaveProperty.class.getName().replaceFirst(".+\\.", ""));
    boolean lastRequestValid = false;
    Object[] lastRequestArgs;
    
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
     * 
     * @param propObj
     * @return 
     */
    @Override
    public boolean validateJSONObject(JSONObject propObj, String type) {      
        return true;
    }
    
    /**
     * In this case, Object args will be OWLClass, OWLDataProperty
     * @param args
     * @return 
     */
    @Override
    public boolean validateRequest(Object... args) {
        if (args.length > 1) {
            OWLClass cls = (OWLClass) args[0];
            OWLDataProperty opd = (OWLDataProperty) args[1];
            OWLReasoner reasoner = Ontology.instance().getReasoner();
            ArrayList invalidInd = new ArrayList();
            boolean valid = true;   
            
            if (!Ontology.instance().getIsConfigurationLoaded()) {
                Ontology.instance().addConfigurationOntology();
            }
            
            for (OWLClass c : Ontology.instance().getClassesInSignature()) {
                if (!c.toString().contains("#") || !cls.toString().contains("#")) continue;
                if (c.toString().split("#")[1].equals(cls.toString().split("#")[1])){
                    NodeSet<OWLNamedIndividual> instances = reasoner.getInstances(c, false);
                    for (OWLNamedIndividual i : instances.getFlattened()) {
                        Set<OWLDataPropertyAssertionAxiom> properties = Ontology.instance().getRootOntology().getDataPropertyAssertionAxioms(i);
                        Set<OWLDataPropertyAssertionAxiom> configProperties = Collections.emptySet();
                        Set<OWLDataPropertyAssertionAxiom> partialProperties = Collections.emptySet();
                        if (Ontology.instance().getIsConfigurationLoaded()) {
                            configProperties = Ontology.instance().getOntology(Ontology.instance().getConfig().getConfigPrefix()).getDataPropertyAssertionAxioms(i);
                        }
                        if (Ontology.instance().getIsPartialConfigurationLoaded()) {
                            configProperties = Ontology.instance().getOntology(Ontology.instance().getConfig().getPartialPrefix()).getDataPropertyAssertionAxioms(i);
                        }
                        Set<OWLDataPropertyAssertionAxiom> policyProperties = Ontology.instance().getOntology(Ontology.instance().getConfig().getPolicyPrefix()).getDataPropertyAssertionAxioms(i);
                        
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
                        
                        for (OWLDataPropertyAssertionAxiom ax: partialProperties) {
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
                            String indStr = i.toString();
                            indStr = indStr.replace("<", "\"");
                            indStr = indStr.replace(">", "\"");
                            String propStr = "\"" + opd.toString().split("#")[1] + "\"";
                            String obj = "{\"individual\":" + indStr + ",\"property\":" + propStr + "}";                     
                            invalidInd.add(obj);
                        }
                    }
                }
            }
            
            if (!valid) {
                explanation = invalidInd.toString();
                lastRequestValid = false;
            } else {
                explanation = "";
                lastRequestValid = true;
                lastRequestArgs = args;
            }
            
            if (args.length > 2) {
                if (args[2].equals("add")) {
                    explanation = "";
                    lastRequestValid = true;
                    lastRequestArgs = args;
                    valid = true;
                }
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
        fullExplanation += ClassMustHaveProperty.class.getName().replaceFirst(".+\\.", "");
        
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
        if (lastRequestValid == false || lastRequestArgs.length == 0) {
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
            IRI classTargetIRI = IRI.create(cwPrefix + "#" + "hasClassTarget");
            IRI requiresPropertyIRI = IRI.create(cwPrefix + "#" + "requiresProperty");

            ind.setProperty(reasonerNameIRI, "ClassMustHaveProperty");

            String className = lastRequestArgs[0].toString().split("#")[1].replace(">", "");
            String requiredPropertyName = lastRequestArgs[1].toString().split("#")[1].replace(">", "");

            ind.setProperty(classTargetIRI, className);
            ind.setProperty(requiresPropertyIRI, requiredPropertyName);
            
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
