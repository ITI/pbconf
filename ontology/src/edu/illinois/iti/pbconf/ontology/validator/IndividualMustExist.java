/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology.validator;

import edu.illinois.iti.pbconf.ontology.ClosedWorld;
import edu.illinois.iti.pbconf.ontology.Individual;
import java.util.Set;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.OWLIndividualAxiom;

/**
 *
 * @author anderson
 */
public class IndividualMustExist implements ClosedWorld.Validator {
    static final Logger logger = Logger.getLogger(IndividualMustExist.class.getName().replaceFirst(".+\\.", ""));
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
     * In this case, Object args will be a single individual.
     * @param args
     * @return 
     */
    @Override
    public boolean validateRequest(Object... args) {
        if (args.length > 0) {
            Individual individual = (Individual) args[0];   
            Set<OWLIndividualAxiom> axioms = individual.getAxioms();
            explanation = "";         
            if (args.length > 1) {
                if (args[1].equals("add")) {
                    return true;
                } else {
                    return axioms.size()>0;
                }
            } else {
                return axioms.size()>0;
            }
        } else {
            explanation = "Invalid argument(s)";
            return false;
        }      
    } 
    
    /**
     * Get an explanation if one exists
     * @return
     */
    @Override
    public String getExplanation() {
        return explanation;      
    }

    @Override
    public boolean saveLastRequest(String nextCWR) {
        throw new UnsupportedOperationException("Not supported yet."); //To change body of generated methods, choose Tools | Templates.
    }
}
