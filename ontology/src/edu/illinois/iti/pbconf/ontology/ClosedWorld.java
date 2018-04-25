/*
 */
package edu.illinois.iti.pbconf.ontology;

import java.lang.reflect.Constructor;
import java.lang.reflect.InvocationTargetException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import org.apache.log4j.Logger;
import org.json.JSONObject;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLDataPropertyExpression;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLObjectProperty;
import org.semanticweb.owlapi.model.OWLObjectPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.util.OWLEntityRemover;

/**
 * This class encompasses the implementation of explicit reasoning that may be 
 * impossible or impractical to infer from the OWL reasoner. Specifically, this
 * will include validations of configuration or policy based on closed world assumptions
 * not expressible in OWL. 
 * 
 * 
 * @author anderson
 */
public class ClosedWorld {
    static final Logger logger = Logger.getLogger(ClosedWorld.class.getName().replaceFirst(".+\\.", ""));
    static String fullExplanation = null;
    
    /**
     * Will have to make the following functions capable of override
     */
    public interface Validator {

        /**
         * Override the class and prefix used for storage in the ontology
         * This lets us setup test CWR's using the TESTClosedWorldReasoner class
         * @param cls
         * @param prefix
         */
        public void overrideClass(String cls, String prefix);

        /**
         * Validate a standard request, varies by function utility
         * @param args
         * @return
         */
        public boolean validateRequest(Object... args);

        /**
         * Validate a json object - used for preprocessed property restriction requests
         * @param propObj
         * @return
         */
        public boolean validateJSONObject(JSONObject propObj, String type);

        /**
         * Save the previous request
         * @param nextCWR
         * @return
         */
        public boolean saveLastRequest(String nextCWR);

        /**
         * Get an explanation in case of a failed CWR
         * @return
         */
        public String getExplanation();

        /**
         * Get the current set class
         * @return
         */
        public String getCurrentClass();

        /**
         * Get the current set prefix
         * @return
         */
        public String getCurrentPrefix();
    }
    
    /**
     * Get an instance of validator for the registered reasoner name.
     * 
     * @param reasonerName
     * @return 
     */
    static public Validator getValidator(String reasonerName) {
        OntologyConfig config = Ontology.instance().getConfig();
        Map<String, String> reasoners = config.getClosedWorldReasoners();
        
        String className = reasoners.get(reasonerName);
        ClassLoader loader = ClosedWorld.class.getClassLoader();
        Validator validator = null;
        try {
            Class reasonerClass = loader.loadClass(className);
            // get an instance of reasoner
            Constructor ctor = reasonerClass.getConstructor();
            validator = (Validator) ctor.newInstance();
        } catch (ClassNotFoundException|SecurityException|NoSuchMethodException|InstantiationException 
                |IllegalAccessException|IllegalArgumentException|InvocationTargetException ex) {
            logger.error("Failed to instantiate "+className, ex);
        } 
        return validator;
    }
    
    /**
     * Helper to parse out xsd int value for comparison
     * @param val
     * @return 
     */
    static public Integer parseXSDInt(String val) {
        String trimmedStr = "";
        if (val.contains("^^")) {
            String[] parts = val.split("\\^\\^");
            trimmedStr = parts[0];
            trimmedStr = trimmedStr.replaceAll("^\"|\"$", "");
        }
        try {
            return Integer.parseInt(trimmedStr);
        } catch (Exception ex) {
            return 0;
        }
    }
    
    /**
     * If we're checking type, we're making sure the prop obj satisfies the axiom
     * axiom can be isA, or isNotA
     * @param propObj
     * @param exp
     * @return 
     */
    static public boolean testType(JSONObject propObj, OWLClassExpression exp) {
        String klass = exp.toString();
        String klassName = klass.split("#")[1].replace(">", "");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLIndividual ind = odf.getOWLNamedIndividual(Ontology.instance().getIRI(propObj.getString("translatedSubject")));
        boolean valid;
        
        switch (propObj.getString("predicate")) {
            case "isA":
                valid = false;
                for (OWLClass cls : ind.getClassesInSignature()) {
                    String clsName = cls.toString().split("#")[1].replace(">", "");
                    if (clsName.equalsIgnoreCase(klassName)) {
                        valid = true;
                    }
                }   
                break;
            case "isNotA":
                valid = true;
                for (OWLClass cls : ind.getClassesInSignature()) {
                    String clsName = cls.toString().split("#")[1].replace(">", "");
                    if (clsName.equalsIgnoreCase(klassName)) {
                        valid = false;
                    }
                }   
                break;
            default:
                //If it's not a type request, we can just return true for this test
                valid = true;
                break;
        }
        
        logger.info("Testing owl class : " + klass + " for propName : " + klassName + " state = " + valid);       

        return valid;
    }
    
    /**
     * If we're checking type, we're making sure the prop obj satisfies the axiom
     * axiom can be isA, or isNotA. Get explanation
     * @param propObj
     * @param exp
     * @return 
     */
    static public String getTypeReason(JSONObject propObj, OWLClassExpression exp) {
        String klass = exp.toString();
        String klassName = klass.split("#")[1].replace(">", "");
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        OWLIndividual ind = odf.getOWLNamedIndividual(Ontology.instance().getIRI(propObj.getString("translatedSubject")));
        boolean valid;
        String expl = "";
        
        switch (propObj.getString("predicate")) {
            case "isA":
                valid = false;
                for (OWLClass cls : ind.getClassesInSignature()) {
                    String clsName = cls.toString().split("#")[1].replace(">", "");
                    if (clsName.equalsIgnoreCase(klassName)) {
                        valid = true;
                    }
                } 
                if (!valid) {
                    expl = "Individual is not of class " + klassName;
                }
                break;
            case "isNotA":
                valid = true;
                for (OWLClass cls : ind.getClassesInSignature()) {
                    String clsName = cls.toString().split("#")[1].replace(">", "");
                    if (clsName.equalsIgnoreCase(klassName)) {
                        valid = false;
                    }
                }  
                if (!valid) {
                    expl = "Individual is of class " + klassName;
                }
                break;
            default:
                //If it's not a type request, we can just return true for this test
                valid = true;
                expl = "";
                break;
        }
        
        return expl;
    }
    
    /**
     * Based on JSON content, test the validity of an object assertion axiom
     * @param propObj
     * @param ax
     * @return 
     */
    static public boolean testObjectProperty(JSONObject propObj, OWLObjectPropertyAssertionAxiom ax) {
        String propName = ax.getProperty().toString().split("#")[1].replace(">", "");
        boolean subValid = true;
        
        if (propName.equals(propObj.getString("translatedObject"))) {
            switch (propObj.getString("predicate")) {
                case "state":
                case "status":    
                    String val = ax.getObject().toString().toLowerCase();
                    if (val.contains("#")) {
                        val = val.split("#")[1].replace(">", "");
                    }
                    String targetVal = propObj.getString("value").toLowerCase();    
                    if (targetVal.contains("#")) {
                        targetVal = targetVal.split("#")[1].replace(">", "");
                    }
                    if (!val.equals(targetVal)) {
                        subValid = false;
                    }
                    break;
                case "min-length":    
                    Integer mlVal = propObj.getInt("value");
                    Integer mltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (mltVal <= mlVal) {
                        subValid = false;
                    }
                    break;
                case "gt": 
                    Integer gtVal = propObj.getInt("value");
                    Integer gttVal = parseXSDInt(ax.getObject().toString());
                    if (gttVal <= gtVal) {
                        subValid = false;
                    }
                    break;
                case "max-length":    
                    Integer maxlVal = propObj.getInt("value");
                    Integer maxltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (maxltVal >= maxlVal) {
                        subValid = false;
                    }
                    break;    
                case "lt":   
                    Integer ltVal = propObj.getInt("value");
                    Integer lttVal = parseXSDInt(ax.getObject().toString());
                    if (lttVal >= ltVal) {
                        subValid = false;
                    }
                    break;   
                case "gte":
                    Integer gteVal = propObj.getInt("value");
                    Integer gtetVal = Integer.parseInt(ax.getObject().toString());
                    if (gtetVal < gteVal) {
                        subValid = false;
                    }
                    break;
                case "lte":
                    Integer lteVal = propObj.getInt("value");
                    Integer ltetVal = parseXSDInt(ax.getObject().toString());
                    if (ltetVal > lteVal) {
                        subValid = false;
                    }
                    break;    
                case "eq":
                    Integer eqVal = propObj.getInt("value");
                    Integer eqtVal = parseXSDInt(ax.getObject().toString());
                    if (!Objects.equals(eqVal, eqtVal)) {
                        subValid = false;
                    }
                    break;    
                case "neq":
                    Integer neqVal = propObj.getInt("value");
                    Integer neqtVal = parseXSDInt(ax.getObject().toString());
                    if (Objects.equals(neqVal, neqtVal)) {
                        subValid = false;
                    }
                    break;
                case "complexity":
                    String compVal =  parseXSDStr(propObj.getString("value"));
                    String compTVal = ax.getObject().toString();
                    if (compVal.toLowerCase().equals("lowercase")) {
                        if (!compTVal.equals(compTVal.toLowerCase())) {
                            subValid = false;
                        }
                    } else if (compVal.toLowerCase().equals("uppercase")) {
                        if (!compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                        }
                    } else if (compVal.toLowerCase().equals("mixedcase")) {
                        if (compTVal.equals(compTVal.toLowerCase()) || compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                        }
                    } else {
                        subValid = false;
                    }
                    break;
                default:
                    break;
            }
        }
        
        return subValid;
    }
    
    /**
     * Based on JSON content, test the validity of an object assertion axiom
     * @param propObj
     * @param ax
     * @return 
     */
    static public String getObjectPropertyReason(JSONObject propObj, OWLObjectPropertyAssertionAxiom ax) {
        String propName = ax.getProperty().toString().split("#")[1].replace(">", "");
        boolean subValid = true;
        String expl = "";
        
        if (propName.equals(propObj.getString("translatedObject"))) {
            switch (propObj.getString("predicate")) {
                case "state":
                case "status":    
                    String val = ax.getObject().toString().toLowerCase();
                    if (val.contains("#")) {
                        val = val.split("#")[1].replace(">", "");
                    }
                    String targetVal = propObj.getString("value").toLowerCase();    
                    if (targetVal.contains("#")) {
                        targetVal = targetVal.split("#")[1].replace(">", "");
                    }
                    if (!val.equals(targetVal)) {
                        subValid = false;
                        expl = val + " state != " + targetVal;
                    }
                    break;
                case "min-length":    
                    Integer mlVal = propObj.getInt("value");
                    Integer mltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (mltVal <= mlVal) {
                        subValid = false;
                        expl = "Too short";
                    }
                    break;
                case "gt": 
                    Integer gtVal = propObj.getInt("value");
                    Integer gttVal = parseXSDInt(ax.getObject().toString());
                    if (gttVal <= gtVal) {
                        subValid = false;
                         expl = "Too short";
                    }
                    break;
                case "max-length":    
                    Integer maxlVal = propObj.getInt("value");
                    Integer maxltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (maxltVal >= maxlVal) {
                        subValid = false;
                         expl = "Too long";
                    }
                    break;    
                case "lt":   
                    Integer ltVal = propObj.getInt("value");
                    Integer lttVal = parseXSDInt(ax.getObject().toString());
                    if (lttVal >= ltVal) {
                        subValid = false;
                         expl = "Too long";
                    }
                    break;   
                case "gte":
                    Integer gteVal = propObj.getInt("value");
                    Integer gtetVal = Integer.parseInt(ax.getObject().toString());
                    if (gtetVal < gteVal) {
                        subValid = false;
                         expl = "Too short";
                    }
                    break;
                case "lte":
                    Integer lteVal = propObj.getInt("value");
                    Integer ltetVal = parseXSDInt(ax.getObject().toString());
                    if (ltetVal > lteVal) {
                        subValid = false;
                         expl = "Too long";
                    }
                    break;    
                case "eq":
                    Integer eqVal = propObj.getInt("value");
                    Integer eqtVal = parseXSDInt(ax.getObject().toString());
                    if (!Objects.equals(eqVal, eqtVal)) {
                        subValid = false;
                        expl = eqVal.toString() + " != " + eqtVal.toString();
                    }
                    break;    
                case "neq":
                    Integer neqVal = propObj.getInt("value");
                    Integer neqtVal = parseXSDInt(ax.getObject().toString());
                    if (Objects.equals(neqVal, neqtVal)) {
                        subValid = false;
                        expl = neqVal.toString() + " == " + neqtVal.toString();
                    }
                    break;
                case "complexity":
                    String compVal =  parseXSDStr(propObj.getString("value"));
                    String compTVal = ax.getObject().toString();
                    if (compVal.toLowerCase().equals("lowercase")) {
                        if (!compTVal.equals(compTVal.toLowerCase())) {
                            subValid = false;
                            expl = "Not lower case";
                        }
                    } else if (compVal.toLowerCase().equals("uppercase")) {
                        if (!compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                            expl = "Not upper case";
                        }
                    } else if (compVal.toLowerCase().equals("mixedcase")) {
                        if (compTVal.equals(compTVal.toLowerCase()) || compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                            expl = "Not mixed case";
                        }
                    } else {
                        subValid = false;
                        expl = "Invalid complexity value";
                    }
                    break;
                default:
                    break;
            }
        }
        
        if (subValid == true) {
            expl = "";
        }
        
        return expl;
    }
    
    /**
     * Test a data property against a prop object
     * @param propObj
     * @param ax
     * @return 
     */
    static public boolean testDataProperty(JSONObject propObj, OWLDataPropertyAssertionAxiom ax) {
        String propertyName = ax.getProperty().toString().split("#")[1].replace(">", "");
        boolean subValid = true;
        
        if (propertyName.equals(propObj.getString("translatedSubject"))) {
            switch (propObj.getString("predicate")) {
                case "state":
                case "status":    
                    String val = ax.getObject().toString();
                    if (val.contains("#")) {
                        val = val.split("#")[1].replace(">", "");
                    }
                    String targetVal = propObj.getString("value").toLowerCase();                 
                    if (!val.equals(targetVal)) {
                        subValid = false;
                    }
                    break;
                case "min-length":    
                    Integer mlVal = propObj.getInt("value");
                    Integer mltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (mltVal <= mlVal) {
                        subValid = false;
                    }
                    break;
                case "gt": 
                    Integer gtVal = propObj.getInt("value");
                    Integer gttVal = parseXSDInt(ax.getObject().toString());
                    if (gttVal <= gtVal) {
                        subValid = false;
                    }
                    break;
                case "max-length":    
                    Integer maxlVal = propObj.getInt("value");
                    Integer maxltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (maxltVal >= maxlVal) {
                        subValid = false;
                    }
                    break;    
                case "lt":   
                    Integer ltVal = propObj.getInt("value");
                    Integer lttVal = parseXSDInt(ax.getObject().toString());
                    if (lttVal >= ltVal) {
                        subValid = false;
                    }
                    break;    
                case "gte":
                    Integer gteVal = propObj.getInt("value");
                    Integer gtetVal = parseXSDInt(ax.getObject().toString());
                    if (gtetVal < gteVal) {
                        subValid = false;
                    }
                    break;
                case "lte":
                    Integer lteVal = propObj.getInt("value");
                    Integer ltetVal = parseXSDInt(ax.getObject().toString());
                    if (ltetVal > lteVal) {
                        subValid = false;
                    }
                    break;    
                case "eq":
                    Integer eqVal = propObj.getInt("value");
                    Integer eqtVal = parseXSDInt(ax.getObject().toString());
                    if (!Objects.equals(eqVal, eqtVal)) {
                        subValid = false;
                    }
                    break;    
                case "neq":
                    Integer neqVal = propObj.getInt("value");
                    Integer neqtVal = parseXSDInt(ax.getObject().toString());
                    if (Objects.equals(neqVal, neqtVal)) {
                        subValid = false;
                    }
                    break;
                case "complexity":
                    String compVal = parseXSDStr(propObj.getString("value"));
                    String compTVal = ax.getObject().toString();
                    if (compVal.toLowerCase().equals("lowercase")) {
                        if (!compTVal.equals(compTVal.toLowerCase())) {
                            subValid = false;
                        }
                    } else if (compVal.toLowerCase().equals("uppercase")) {
                        if (!compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                        }
                    } else if (compVal.toLowerCase().equals("mixedcase")) {
                        if (compTVal.equals(compTVal.toLowerCase()) || compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                        }
                    } else {
                        subValid = false;
                    }
                    break;
                default:
                    break;
            }
        }
        
        return subValid;
    }
    
    /**
     * Test a data property against a prop object, get explanation
     * @param propObj
     * @param ax
     * @return 
     */
    static public String getDataPropertyReason(JSONObject propObj, OWLDataPropertyAssertionAxiom ax) {
        String propertyName = ax.getProperty().toString().split("#")[1].replace(">", "");
        boolean subValid = true;
        String expl = "";
        
        if (propertyName.equals(propObj.getString("translatedSubject"))) {
            switch (propObj.getString("predicate")) {
                case "state":
                case "status":    
                    String val = ax.getObject().toString();
                    if (val.contains("#")) {
                        val = val.split("#")[1].replace(">", "");
                    }
                    String targetVal = propObj.getString("value").toLowerCase();                 
                    if (!val.equals(targetVal)) {
                        subValid = false;
                        expl = val + " state != " + targetVal;
                    }
                    break;
                case "min-length":    
                    Integer mlVal = propObj.getInt("value");
                    Integer mltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (mltVal <= mlVal) {
                        subValid = false;
                         expl = "Too short";
                    }
                    break;
                case "gt": 
                    Integer gtVal = propObj.getInt("value");
                    Integer gttVal = parseXSDInt(ax.getObject().toString());
                    if (gttVal <= gtVal) {
                        subValid = false;
                         expl = "Too short";
                    }
                    break;
                case "max-length":    
                    Integer maxlVal = propObj.getInt("value");
                    Integer maxltVal = parseXSDStr(ax.getObject().toString()).length();
                    if (maxltVal >= maxlVal) {
                        subValid = false;
                         expl = "Too long";
                    }
                    break;    
                case "lt":   
                    Integer ltVal = propObj.getInt("value");
                    Integer lttVal = parseXSDInt(ax.getObject().toString());
                    if (lttVal >= ltVal) {
                        subValid = false;
                         expl = "Too long";
                    }
                    break;    
                case "gte":
                    Integer gteVal = propObj.getInt("value");
                    Integer gtetVal = parseXSDInt(ax.getObject().toString());
                    if (gtetVal < gteVal) {
                        subValid = false;
                         expl = "Too short";
                    }
                    break;
                case "lte":
                    Integer lteVal = propObj.getInt("value");
                    Integer ltetVal = parseXSDInt(ax.getObject().toString());
                    if (ltetVal > lteVal) {
                        subValid = false;
                         expl = "Too long";
                    }
                    break;    
                case "eq":
                    Integer eqVal = propObj.getInt("value");
                    Integer eqtVal = parseXSDInt(ax.getObject().toString());
                    if (!Objects.equals(eqVal, eqtVal)) {
                        subValid = false;
                        expl = eqVal.toString() + " != " + eqtVal.toString();
                    }
                    break;    
                case "neq":
                    Integer neqVal = propObj.getInt("value");
                    Integer neqtVal = parseXSDInt(ax.getObject().toString());
                    if (Objects.equals(neqVal, neqtVal)) {
                        subValid = false;
                        expl = neqVal.toString() + " == " + neqtVal.toString();
                    }
                    break;
                case "complexity":
                    String compVal = parseXSDStr(propObj.getString("value"));
                    String compTVal = ax.getObject().toString();
                    if (compVal.toLowerCase().equals("lowercase")) {
                        if (!compTVal.equals(compTVal.toLowerCase())) {
                            subValid = false;
                            expl = "Not lower case";
                        }
                    } else if (compVal.toLowerCase().equals("uppercase")) {
                        if (!compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                            expl = "Not upper case";
                        }
                    } else if (compVal.toLowerCase().equals("mixedcase")) {
                        if (compTVal.equals(compTVal.toLowerCase()) || compTVal.equals(compTVal.toUpperCase())) {
                            subValid = false;
                            expl = "Not mixed case";
                        }
                    } else {
                        subValid = false;
                        expl = "Invalid complexity value";
                    }
                    break;
                default:
                    break;
            }
        }
        
        if (subValid == true) {
            expl = "";
        }
        
        return expl;
    }
    
    /**
     * Call reason with default values (called this way outside of testing generally)
     * @return 
     */
    static public boolean reason(boolean isTest) {
        String defaultValidatorClass = "ClosedWorldReasoner";
        String defaultValidatorIndPrefix = "cwr";
        if (isTest) {
            defaultValidatorClass = "TESTClosedWorldReasoner";
            defaultValidatorIndPrefix = "tcwr";
        }       
        return reason(defaultValidatorClass, defaultValidatorIndPrefix);
    }
    
    /**
     * Determine if we're currently consistent based on all closed world rules
     * Unfortunately we'll have to go through all of them each time.  We can 
     * store them in memory for faster operations.
     * @param cls
     * @param prefix
     * @return 
     */
    static public boolean reason(String cls, String prefix) {
        ArrayList<HashMap<String, String>> cwrs = getClosedWorldReasoners(cls);
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        //Start with a valid state and work through each of the closed world reasoners
        //Each result will be combined with current state.  If he become invalid
        //we can stop there, and report
        boolean valid = true;
        String explanation = "";
        for (HashMap<String, String> cwr : cwrs) {
            //Get the validator associated with this closed world reasoner
            //process based on values in HashMap and requirements of reasoner
            String reasonerName = cwr.get("hasReasonerName");
            ClosedWorld.Validator validator = ClosedWorld.getValidator(reasonerName);
            
            if ("ClassMustHaveProperty".equals(reasonerName) || "ClassMustNotHaveProperty".equals(reasonerName)) {
                OWLClass ocls = odf.getOWLClass(Ontology.instance().getIRI(cwr.get("hasClassTarget")));
                OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI(cwr.get("requiresProperty")));
                
                boolean isValid = validator.validateRequest(ocls, odp, "validate");
                valid = valid && isValid;
                
                if (!isValid) {
                    fullExplanation = validator.getExplanation();
                    logger.info("Failed to process ClosedWorld reasoning with explanation : ");
                    logger.info(explanation);
                    //Make sure we break so as to get only the first explanation
                    break;
                }
            }
            
            if ("IndividualMustHaveProperty".equals(reasonerName) || "IndividualMustNotHaveProperty".equals(reasonerName)) {
                OWLIndividual ind = odf.getOWLNamedIndividual(Ontology.instance().getIRI(cwr.get("hasIndividualTarget")));
                OWLDataProperty odp = odf.getOWLDataProperty(Ontology.instance().getIRI(cwr.get("requiresProperty")));
                
                boolean isValid = validator.validateRequest(ind, odp, "validate");
                valid = valid && isValid;
                
                if (!isValid) {
                    fullExplanation = validator.getExplanation();
                    logger.info("Failed to process ClosedWorld reasoning with explanation : ");
                    logger.info(explanation);
                    //Make sure we break so as to get only the first explanation
                    break;
                }
            }
            
            if ("ClassPropertyRestrictions".equals(reasonerName) || "IndividualPropertyRestrictions".equals(reasonerName)) {
                JSONObject jObj = ClosedWorld.generateJSON(cwr);
                
                boolean isValid = validator.validateJSONObject(jObj, "validate");
                valid = valid && isValid;
                
                if (!isValid) {
                    fullExplanation = validator.getExplanation();
                    logger.info("Failed to process ClosedWorld reasoning with explanation : ");
                    logger.info(explanation);
                    //Make sure we break so as to get only the first explanation
                    break;
                }
            }
        }
        
        if (valid) {
            fullExplanation = "";
        }
        
        return valid;
    }
    
    /**
     * Helper function to rebuild json object after storing in database
     * @param cwr
     * @return 
     */
    static JSONObject generateJSON(HashMap<String, String> cwr) {
        String subject = cwr.get("subject");
        String predicate = cwr.get("predicate");
        String object = cwr.get("object");
        String target = "";
        switch (cwr.get("hasReasonerName")) {
            case "ClassPropertyRestrictions": {
                target = cwr.get("hasClassTarget");
                break;
            }
            case "IndividualPropertyRestrictions":{
                target = cwr.get("hasIndividualTarget");
                break;
            }
            default:{
                break;
            }
        }
        
        PBConfParser pbParser = new PBConfParser();
        return pbParser.getAxiomParts(target, subject, predicate, object);
    }
    
    /**
     * Get an explanation of the last issue (empty string if we were valid)
     * @return 
     */
    static String getExplanation() {
        return fullExplanation;
    }
    
    /**
     * Get all closed world reasoners in the policy ontology
     * @return 
     */
    static private ArrayList<HashMap<String, String>> getClosedWorldReasoners(String cls) {
        OntologyConfig cfg = Ontology.instance().getConfig();
        Set<OWLNamedIndividual> inds = Ontology.instance().getOntology(cfg.get("closedWorldOntology")).getIndividualsInSignature();
        ArrayList<HashMap<String, String>> cwrs = new ArrayList<>();
        
        //Get an ArrayList of all closed world reasoners
        for (OWLNamedIndividual ind : inds) {
            if (isClosedWorldReasoner(ind, cls)) {
                cwrs.add(getClosedWorldReasoner(ind));
            }
        }
        
        return cwrs;
    }
    
    /**
     * Get the next available closed world reasoner name
     * @param cls
     * @param prefix
     * @return 
     */
    static public String getNextClosedWorldReasonerName(String cls, String prefix) {
        ArrayList<HashMap<String, String>> cwrs = getClosedWorldReasoners(cls);
        ArrayList<Integer> ids = new ArrayList();
        
        for (HashMap<String, String> cwr : cwrs) {
            String id = cwr.get("id");
            ids.add(Integer.parseInt(id.replace(prefix, "")));
        }
        
        //Fallback in case this is the first closed world reasoner
        if (ids.isEmpty()) {
            return prefix + "0";
        }
        
        Collections.sort(ids);
        Integer last = ids.get(ids.size() - 1) + 1;
        String newName = prefix + last.toString();
        
        return newName;
    }
    
    /**
     * Helper function to determine if 
     * @param clss
     * @return 
     */
    static private boolean isClosedWorldReasoner(OWLNamedIndividual ind, String cls) {
        OntologyConfig cfg = Ontology.instance().getConfig();
        Set<OWLClassExpression> oces = ind.getTypes(Ontology.instance().getOntology(cfg.get("closedWorldOntology")));
        for (OWLClassExpression oce : oces) {
            String oceStr = oce.toString().split("#")[1].replace(">", "");
            if (oceStr.equals(cls)) {
                return true;
            }
        }
        return false;
    }
       
    /**
     * Helper fn for saving / loading
     * @param IRI
     * @return 
     */
    static private String strFromIRI(String IRI) {
        String result = "";
        if (IRI.contains("#")) {
            result = IRI.split("#")[1];
            result = result.replace(">", "");
        }
        return result;
    }
    
    /**
     * Get a HashMap containing ClosedWorldReasoning data 
     * @param ind
     * @return 
     */
    static private HashMap<String, String> getClosedWorldReasoner(OWLNamedIndividual ind) {
        //We know we have a ClosedWorldReasoner, we'll need to get its data properties and reason across it
        HashMap<String, String> curCWR = new HashMap<>();
        OntologyConfig cfg = Ontology.instance().getConfig();

        Set<OWLDataPropertyAssertionAxiom> dps = Ontology.instance().getOntology(cfg.get("closedWorldOntology")).getDataPropertyAssertionAxioms(ind);
        for (OWLDataPropertyAssertionAxiom dp : dps) {
            //getProperty returns key
            OWLDataPropertyExpression odpex = dp.getProperty();

            //literal returns the value (in this case, xsd:string)
            OWLLiteral oLit = dp.getObject();
            String oTest = odpex.toString();
            oTest = strFromIRI(oTest);
            
            curCWR.put(oTest, parseXSDStr(oLit.toString()));

            /*
            if (oTest.contains("hasClassTarget")) {
                curCWR.put("hasClassTarget", parseXSDStr(oLit.toString()));
            } else if (oTest.contains("hasIndividualTarget")) {                
                curCWR.put("hasIndividualTarget", parseXSDStr(oLit.toString()));
            } else if (oTest.contains("requiresProperty")) {
                curCWR.put("requiresProperty", parseXSDStr(oLit.toString()));
            } else if (oTest.contains("hasReasonerName")) {
                curCWR.put("hasReasonerName", parseXSDStr(oLit.toString()));
            }
            */
        }
        
        curCWR.put("id", ind.toString().split("#")[1].replace(">", ""));
        
        return curCWR;
    }
    
    /**
     * Get a cleaned up String value out of the OWL Literal String
     * // ??? this seems scary. 
     * @param litStr
     * @return 
     */
    static private String parseXSDStr(String litStr) {
        String XSDSubStr = "\"^^xsd:string";
        if (litStr.contains(XSDSubStr) && litStr.startsWith("\"")) {
            String resultStr = litStr.replaceFirst("\"", "");
            resultStr = resultStr.replace(XSDSubStr, "");
            return resultStr;
        }
        return litStr;
    }
    
    /**
     * Remove all test reasoners
     * This is a helper function to be run at the start of related unit tests
     * @param cls
     * @param prefix
     * @return 
     */
    static public boolean removeTestReasoners(String cls, String prefix) {
        OntologyConfig cfg = Ontology.instance().getConfig();
        Set<OWLNamedIndividual> inds = Ontology.instance().getOntology(cfg.get("closedWorldOntology")).getIndividualsInSignature();
        OWLOntology pOnt = Ontology.instance().getOntology(cfg.get("closedWorldOntology"));
        OWLOntologyManager oMgr = Ontology.instance().getManager();
        
        OWLEntityRemover er = new OWLEntityRemover(oMgr, Collections.singleton(pOnt));
        for (OWLNamedIndividual ind : inds) {
            if (isClosedWorldReasoner(ind, cls)) {
                er.visit(ind);
                oMgr.applyChanges(er.getChanges());
            }
        }
        
        return true;
    }
    
    /**
     * Clear duplicate reasoners based on class prefix
     * This will check for a match of data properties and object properties between individuals
     * It also requires the individuals be of a class type and have name containing prefix.
     * This is primarily used to cleanup after tests so that we don't get tons of duplicates
     * @param cls 
     * @param prefix 
     * @return  
     */
    static public boolean clearDuplicateClosedWorldReasoners(String cls, String prefix) {
        OntologyConfig cfg = Ontology.instance().getConfig();
        Set<OWLNamedIndividual> inds = Ontology.instance().getOntology(cfg.get("closedWorldOntology")).getIndividualsInSignature();
        OWLOntology pOnt = Ontology.instance().getOntology(cfg.get("closedWorldOntology"));
        OWLOntologyManager oMgr = Ontology.instance().getManager();
        
        //we'll keep a set of unique individuals to compare to
        Set<OWLNamedIndividual> uniqueInds = new HashSet<>();
        //Use the entity remove to clean things up
        OWLEntityRemover er = new OWLEntityRemover(oMgr, Collections.singleton(pOnt));
        
        for (OWLNamedIndividual ind : inds) {
            //Only work on closed world reasoners matching the class we expect
            if (isClosedWorldReasoner(ind, cls)) {
                if (uniqueInds.isEmpty()) {
                    uniqueInds.add(ind);
                } else {
                    boolean found = false;
                    
                    //Get data property assertions and parse them
                    Set<OWLDataPropertyAssertionAxiom> indProps = pOnt.getDataPropertyAssertionAxioms(ind);
                    Set<String> indPropStrs = new HashSet<>();                   
                    for (OWLDataPropertyAssertionAxiom indProp : indProps) {
                        String iProp = indProp.toString();
                        //Here we make sure to make the identifier generic so that doesn't interfere
                        if (iProp.contains("#" + prefix)) {
                            String base = "#" + prefix;
                            iProp = iProp.replaceAll(base + "[0-9]+", "#" + prefix);
                        }
                        indPropStrs.add(iProp);
                    }
                    
                    Set<OWLObjectProperty> indObjProps = ind.getObjectPropertiesInSignature();
                    String indObjPropsStr = indObjProps.toString();
                    
                    for (OWLNamedIndividual uniqueInd : uniqueInds) {
                        Set<OWLDataPropertyAssertionAxiom> uniqueProps = pOnt.getDataPropertyAssertionAxioms(uniqueInd);
                        Set<OWLObjectProperty> uniqueObjProps = uniqueInd.getObjectPropertiesInSignature();
                        String uniqueObjPropsStr = uniqueObjProps.toString();
                        
                        //Same operations on the existing list of unique items
                        Set<String> uniquePropStrs = new HashSet<>();                   
                        for (OWLDataPropertyAssertionAxiom uniqueProp : uniqueProps) {
                            String iProp = uniqueProp.toString();
                            if (iProp.contains("#" + prefix)) {
                                String base = "#" + prefix;
                                iProp = iProp.replaceAll(base + "[0-9]+", "#" + prefix);
                            }
                            uniquePropStrs.add(iProp);
                        }
                        
                        //If we match in property sizes, and match on a per item basis
                        //and match in object properties, we can claim we have a duplicate and remove it
                        if (indPropStrs.size() == uniquePropStrs.size()) {
                            Integer fc = 0;
                            for (String indPropStr : indPropStrs) {
                                for (String uniquePropStr : uniquePropStrs) {
                                    if (indPropStr.equals(uniquePropStr)) {
                                        fc++;
                                    }
                                }
                            }
                            if (fc >= indPropStrs.size() && indObjPropsStr.equals(uniqueObjPropsStr)) {
                                found = true;
                            }
                        }
                    }
                    
                    if (found == true) {
                        er.visit(ind);
                        oMgr.applyChanges(er.getChanges());
                    } else {
                        uniqueInds.add(ind);
                    }
                }
            }
        }
        
        return true;
        
        //Save the duplicate removal back to the closed world ontology bucket
        //Note : no longer saving anything to physical file
        /*
        try {
            Ontology.instance().saveOntology(pOnt, "owl/pbconf." + cfg.get("closedWorldOntology") + ".owl");
            return true;
        } catch (OWLOntologyStorageException | FileNotFoundException ex) {
            logger.info(ex.toString());
            return false;
        }
        */
    }   
}
