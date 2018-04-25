/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.apache.log4j.Logger;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLAxiom;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassExpression;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLEntity;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLObjectComplementOf;
import org.semanticweb.owlapi.model.OWLObjectIntersectionOf;
import org.semanticweb.owlapi.model.OWLObjectUnionOf;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;

/**
 * Used to process equivalentTo and disjointWith requests that allow a more flexible string to be passed
 * as opposed to the standard methods.
 * @author josephdigiovanna
 */
public class OntologyPolicyStringEngine {
    //Handles whether or not we need to add closed world reasoners under a specific test prefix.
    private String axiomStr = "";
    private String reducedStr = "";
    private String classTarget = "";
    private String operation = "";
    private String individualTarget = "";
    private OWLAxiom resultingAxiom = null;
    private String explanation = "";
    private boolean valid;
    private int count = 0;
    private final Map<String, String> valueMap = new HashMap<>();
    private final Map<String, OWLClassExpression> expressionMap = new HashMap<>();
    static Logger logger = Logger.getLogger(OntologyPolicyStringEngine.class.getName().replaceFirst(".+\\.", ""));
    
    /**
     * Setup empty data with default constructor
     */
    public OntologyPolicyStringEngine() {
        this.classTarget = "";
        this.operation = "";
        this.axiomStr = "";
        this.reducedStr = "";
        this.individualTarget = "";
        this.resultingAxiom = null;
        this.valueMap.clear();
        this.expressionMap.clear();
        this.count = 0;
        this.valid = false;
        this.explanation = "";
        
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
    }
    
    /**
     * Allows an axiom string to be used to construct a disjoint axiom for a class.  As an example,
     * you could say (for device type LINUX) "$not ((authDNP $value Joe) $and (authDNP $value Andy))", 
     * which basically says you aren't allowed to have Joe and Andy both have access to LINUX boxes on over DNP3
     * @param cls
     * @param operation
     * @param axiomStr 
     * @return  
     */
    public OWLAxiom constructAxiomFromString(String cls, String operation, String axiomStr) {
        this.classTarget = cls;
        this.operation = operation;
        this.axiomStr = axiomStr;
        this.reducedStr = "";
        this.individualTarget = "";
        this.resultingAxiom = null;
        this.valueMap.clear();
        this.expressionMap.clear();
        this.count = 0;
        this.valid = true;
        this.explanation = "";
        
        if (!Ontology.instance().getIsConfigurationLoaded()) {
            Ontology.instance().addConfigurationOntology();
        }
        
        reduceAxiomStr();
        parseAxiomStr();
        return generateAxiom();
    }

    /**
     * This goal of this is to take any values that aren't key words and reduce them down to 
     * 2 character strings, that way everything is uniform and can be mapped easier.  Also fixes spacing
     * and several other possible formatting issues, leading up to processing
     */
    public void reduceAxiomStr() {
        //First we trim surrounding spaces, then remove extra spaces
        String tmpStr = this.axiomStr;
        
        //Make sure we setup extra spaces around parenthesis so we don't combine terms
        tmpStr = tmpStr.replaceAll("\\(", " \\( ").replaceAll("\\)", " \\) ");
        
        //Remove cases where we've caused more than one consecutive space
        tmpStr = tmpStr.replaceAll(" +", " ");
        
        //Now remove parenthesis
        tmpStr = tmpStr.replaceAll("\\(", "").replaceAll("\\)", "");
        
        //Remove all special relational terms
        tmpStr = tmpStr.replaceAll("\\$and", "").replaceAll("\\$nand", "");
        tmpStr = tmpStr.replaceAll("\\$or", "").replaceAll("\\$nor", "");
        tmpStr = tmpStr.replaceAll("\\$not", "");
        
        //Remove all predicate terms
        tmpStr = tmpStr.replaceAll("\\$value", "").replaceAll("\\$some", "");
        
        //Finally, trim spaces again and remove beginning and ending spaces
        tmpStr = tmpStr.trim().replaceAll(" +", " ");
        
        String[] parts = tmpStr.split(" ");
        
        for (String part : parts) {
            String equivStr = convertCount(this.count);
            this.valueMap.put(equivStr, part);
            this.axiomStr = this.axiomStr.replaceFirst(part, equivStr);
            this.count++;
        }
        
        //Cleanup the actual axiom string
        this.axiomStr = this.axiomStr.replaceAll("\\(", " \\( ").replaceAll("\\)", " \\) ");
        this.axiomStr = this.axiomStr.trim().replaceAll(" +", " ");
    }
    
    /**
     * Parse the reduced into a set of OWLClassExpressions, and then reduce to a single OWLClassExpression
     */
    public void parseAxiomStr() {
        //First we parse out any instances of parenthesis in the string 
        int leftCount = this.axiomStr.length() - this.axiomStr.replace("(", "").length();
        int rightCount = this.axiomStr.length() - this.axiomStr.replace(")", "").length();
        //Make sure we have valid parenthesis counts
        if (leftCount != rightCount) {
            this.valid = false;
        }
        
        //First evaluate everything in parenthesis
        while ((this.axiomStr.contains("(") || this.axiomStr.contains(")")) && valid == true) {
            //Make sure we don't have uneven parenthesis
            if ((this.axiomStr.contains("(") && !this.axiomStr.contains(")")) || (!this.axiomStr.contains("(") && this.axiomStr.contains(")"))) {
                this.valid = false;
                break;
            }
            
            int counter = 0;
            int start = -1;
            int end = -1;
            for (int i = 0; i < this.axiomStr.length(); i++) {
                if ("(".equals(this.axiomStr.substring(i, i+1))) {
                    counter++;
                    start = i;
                }
                if (")".equals(this.axiomStr.substring(i, i+1)) && counter > 0) {
                    counter--;
                    end = i;
                    
                    String substr = this.axiomStr.substring(start+1, end);
                    substr = substr.trim().replaceAll(" +", " ");   
                    String result = parseAxiomSubstring(substr);
                    
                    //Now we replace the parenthesis + interior with a marking character
                    if (start == 0) {
                        this.axiomStr = result + this.axiomStr.substring(end + 1, this.axiomStr.length());
                    } else if (end == this.axiomStr.length() - 1) {
                        this.axiomStr = this.axiomStr.substring(0, start) + result;
                    } else {
                        this.axiomStr = this.axiomStr.substring(0, start) + result + this.axiomStr.substring(end + 1, axiomStr.length());
                    }
                    
                    //Break out and start from beginning
                    i = axiomStr.length() + 1;
                }
            }
        }
        
        //Now that we have no more parenthesis, just call it on the remaining string
        this.axiomStr = parseAxiomSubstring(this.axiomStr);
    }
    
    /**
     * Parse an axiom substring
     * @param substr 
     */
    private String parseAxiomSubstring(String substr) {
        //This has no parenthesis, so we first process terms for expressions {$some, $value}
        //then we process relational terms {$not, $and, $or}, and reduce to a single expression
        substr = substr.trim().replaceAll(" +", " ");
        List<String> parts = new ArrayList<>(Arrays.asList(substr.split(" ")));
        
        //We should have a series of statements related by and,or,not,nand,nor,some,value
        //First we process some / value triples, then we process relators in order precedence
        //not, and, or
        //We will now use these to create relations from the expressions
        //We can split this into an array to make things a bit easier since they're evaluated left to right

        List<String> keywords = new ArrayList();
        keywords.add("$some");
        keywords.add("$value");
        
        List<String> relators = new ArrayList();       
        relators.add("$not");
        relators.add("$and");
        //relators.add("$nand");
        relators.add("$or");       
        //relators.add("$nor");
      
        parts = processKeywords(parts, keywords);     
        parts = processRelators(parts, relators);

        return (String) parts.get(0);
    }
    
    /**
     * Process a list of keywords from an axiom parts list
     * keywords are currently {$some, $value}
     * @param axiomParts
     * @param keywords
     * @return 
     */
    private List processKeywords(List<String> axiomParts, List<String> keywords) {
        boolean success = false;
        for (String keyword : keywords) {
            while (axiomParts.contains(keyword)) {
                int index = axiomParts.indexOf(keyword);
                if (index == 0 || index == keywords.size()) {
                    logger.error("Error processing keywords list, relator was not between terms");
                    return axiomParts;
                } else {
                    String[] pArr = new String[3];
                    pArr[0] = (String)axiomParts.get(index - 1);
                    pArr[1] = (String)axiomParts.get(index); 
                    pArr[2] = (String)axiomParts.get(index + 1);
                    success = buildAxiomFromParts(pArr);
                    
                    if (!success) {
                        this.valid = false;
                        break;
                    }

                    axiomParts.remove(index + 1);
                    axiomParts.remove(index);
                    axiomParts.set(index - 1, convertCount(this.count));
                    this.count++;
                }
            }
        }
        return axiomParts;
    }
    
    /**
     * Process a list of relators from an axiom parts list
     * @param axiomParts
     * @param relators
     * @return 
     */
    private List processRelators(List<String> axiomParts, List<String> relators) {
        //Arraylist order is guaranteed to be consistent, so the initial order
        //set will be considered order precedence.  Should be not, and, or.
        for (String relator : relators) {
            axiomParts = processRelator(axiomParts, (String) relator);
        }
        
        return axiomParts;
    }
    
    /**
     * process a single relator string from an axiom parts list
     * @param axiomParts
     * @param relator
     * @return 
     */
    private List processRelator(List<String> axiomParts, String relator) {
        if (relator.equals("$not")) {
             //As long as we process right to left, we don't have to worry about multiple consecutive terms
            while (axiomParts.contains(relator)) {
                int index = axiomParts.lastIndexOf(relator);
                if (index == axiomParts.size() - 1) {
                    logger.error("Unable to process relator $not due to matching last list item");
                    return axiomParts;
                } else {
                    String[] pArr = new String[2];
                    pArr[0] = axiomParts.get(index); 
                    pArr[1] = axiomParts.get(index + 1);
                    boolean success = buildAxiomFromParts(pArr);

                    if (!success) {
                        this.valid = false;
                        this.explanation = "Unable to build axiom from " + pArr[0] + " " + pArr[1]; 
                        return axiomParts;
                    }

                    axiomParts.remove(index + 1);
                    axiomParts.set(index, convertCount(this.count));
                    this.count++;
                }
            }
        } else {
            while (axiomParts.contains(relator)) {
                //It's an "$and" or "$or" relator, so we're proecessing one or more for each, so 
                //we can cover situations like 'a or b or c and d and e'
                String aParts = joinList(axiomParts);
                String pattern = "\\w+( \\" + relator + " \\w+)+";
                Pattern r = Pattern.compile(pattern);
                Matcher m = r.matcher(aParts);

                if (m.find()) {
                    //Since this is a greedy pattern, should only use the first one and then search again
                    String[] groupParts = m.group(0).split(" ");
                    int startingIndex = findStartingIndex(groupParts, axiomParts);
                    boolean success = buildAxiomFromParts(groupParts);

                    if (!success) {
                        this.valid = false;
                        this.explanation = "Unable to build axiom from ";
                        for (String groupPart : groupParts) {
                            this.explanation += " " + groupPart;
                        }
                        return axiomParts;
                    }

                    for (int i = groupParts.length - 1; i > 0; i--) {
                        axiomParts.remove(startingIndex + i);
                    }
                    
                    axiomParts.set(startingIndex, convertCount(this.count));
                    this.count++;
                } else {
                    logger.error("Found an instance of " + relator + " but pattern didn't match");
                    break;
                }
            }
        }
        
        return axiomParts;
    }
    
    /**
     * Given a string array (matched substring), find it in the axiomParts array list, returning starting index
     * @param groupParts
     * @param axiomParts
     * @return 
     */
    private int findStartingIndex(String[] groupParts, List<String> axiomParts) {
        boolean found = false;
        int index = -1;
        for (int i = 0; i < axiomParts.size(); i++) {
            if (axiomParts.get(i).equals(groupParts[0]) && ((axiomParts.size() - i) >= groupParts.length)) {
                boolean subfound = true;
                for (int j = 1; j < groupParts.length; j++) {
                    if (!groupParts[j].equals(axiomParts.get(i + j))) {
                        subfound = false;
                        break;
                    }
                }
                if (subfound == true) {
                    found = true;
                    index = i;
                    break;
                }
            }
        }
        
        return index;
    }
        
    /**
     * Helper function to join a list with spaces
     * @param parts
     * @return 
     */
    private String joinList(List<String> parts) {
        String result = "";
        for (int i = 0; i < parts.size(); i++) {
            if (i != 0) {
                result += " ";
            }
            result = result.concat((String)parts.get(i));
        }
        return result;
    }
    
    /**
     * Generate and return an OWLAxiom from the final expression (should be a combination of all expressions)
     * Return null if there was an error processing the statement
     * @return 
     */
    public OWLAxiom generateAxiom() {
        if (this.valid == false) {
            return null;
        }
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        Set <OWLClassExpression> exps = new HashSet<>();
        OWLClass owlCls = odf.getOWLClass(Ontology.instance().getIRI(this.classTarget));
        OWLClassExpression exp;
        String cc = convertCount(this.count - 1);
        if (this.expressionMap.containsKey(cc)) {
            exp = this.expressionMap.get(cc);
        } else if (this.valueMap.containsKey(cc)) {
            exp = odf.getOWLClass(Ontology.instance().getIRI(this.valueMap.get(cc)));
        } else {
            logger.error("Invalid mapping");
            return null;
        }
        
        exps.add(owlCls);
        exps.add(exp);  
        OWLAxiom ax = null;
        switch (this.operation) {
            case "disjointWith":
                ax = odf.getOWLDisjointClassesAxiom(exps);
                break;
            case "equivalentTo":
                ax = odf.getOWLEquivalentClassesAxiom(exps);
                break;
            default:
                ax = null;
                break;
        }
        
        if (ax == null) {
            logger.error("Unable to generate Axiom");
            return null;
        } else {
            return ax;
        }
    }
    
    /**
     * Convert an integer into a 2 character string consisting of lowercase, and uppercase
     * aa = 0;
     * ZZ = 2703; //2704 - 1
     * //We can expand to 3 characters if need be, but this should cover all potential cases
     * @param value
     * @return 
     */
    public String convertCount(int value) {
        if (value < 0) {
            return "aa";
        }
        if (value > 2703) {
            return "ZZ";
        }
        String convertStr = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
        int cLen = convertStr.length();
        int quotient = Math.floorDiv(value, cLen);
        int remainder = value % cLen;
        String partA = convertStr.substring(quotient, quotient + 1);
        String partB = convertStr.substring(remainder, remainder + 1);
        return partA + partB;
    }
    
    /**
     * Get the location of the first keyword found
     * @param keywords
     * @return 
     */
    public int firstKeywordInt(String[] keywords) {
        boolean found = false;
        int earliest = 32767;
        
        for (String keyword : keywords) {
            if (this.axiomStr.contains(keyword)) {
                int start = this.axiomStr.indexOf(keyword);
                if (start < earliest) {
                    earliest = start;
                }
            }
            
        }
               
        if (found == false) {
            return -1;
        } else {
            return earliest;
        }
    }
    
    /**
     * Get the value of the first keyword found
     * @param start
     * @return 
     */
    public String firstKeywordStr(int start) {
        if (this.axiomStr.length() <= (start + 3)) {
            if (this.axiomStr.length() <= (start + 2)) {
                return "";
            } else {
                if ("$or".equals(this.axiomStr.substring(start, 3))) {
                    return "$or";
                } else {
                    return "";
                }
            }
        } else {
            if ("$and".equals(this.axiomStr.substring(start, start + 4))) {
                return "$and";
            } else if ("$not".equals(this.axiomStr.substring(start, start + 4))) {
                return "$not";
            } else if ("$nor".equals(this.axiomStr.substring(start, start + 4))) {
                return "$nor";
            } else if ("$nand".equals(this.axiomStr.substring(start, start + 5))) {
                return "$nand";
            } else {
                return "";
            }
        }
    }
    
    /**
     * Build a relation between a series of terms and a relator type {$and, $or}
     * @param parts
     * @return 
     */
    public OWLClassExpression buildRelationalRestriction(String[] parts) {
        String relator = parts[1];
        //Can only work with specific relator options
        if (!relator.equals("$or") && !relator.equals("$and")) {
            return null;
        }
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        Set<OWLClassExpression> set = new HashSet<>();
        for (int i = 0; i < parts.length; i++) {
            if (i % 2 == 0) {
                if (this.expressionMap.containsKey(parts[i])) {
                    set.add(this.expressionMap.get(parts[i]));
                }
            }
        }
        switch (relator) {
            case "$and":
                OWLObjectIntersectionOf intersection = odf.getOWLObjectIntersectionOf(set);
                return intersection;
            case "$or":
                OWLObjectUnionOf union = odf.getOWLObjectUnionOf(set);
                return union;
            default:
                return null;    
        }
    }
    
    /**
     * Build a relation between two axioms ($and, $or, $nand, $nor) and return
     * @param relator
     * @param axa
     * @param axb
     * @return 
     */
    public OWLClassExpression buildRelationalRestriction(String relator, OWLClassExpression axa, OWLClassExpression axb) {
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        if (null != relator) switch (relator) {
            case "$and":
                OWLObjectIntersectionOf intersection = odf.getOWLObjectIntersectionOf(axa, axb);
                return intersection;
            case "$or":
                OWLObjectUnionOf union = odf.getOWLObjectUnionOf(axa, axb);
                return union;
            case "$nand":
                OWLObjectIntersectionOf nIntersection = odf.getOWLObjectIntersectionOf(axa, axb);
                OWLObjectComplementOf cnIntersection = odf.getOWLObjectComplementOf(nIntersection);
                return cnIntersection;
            case "$nor":
                OWLObjectUnionOf nUnion = odf.getOWLObjectUnionOf(axa, axb);
                OWLObjectComplementOf cnUnion = odf.getOWLObjectComplementOf(nUnion);
                return cnUnion;    
            default:
                return null;
        }
        
        return null;
    }
    
    /**
     * Build a relation on an axiom ($not) and return
     * @param relator
     * @param axa
     * @return 
     */
    public OWLClassExpression buildRelationalRestriction(String relator, OWLClassExpression axa) {
        OWLDataFactory odf = Ontology.instance().getDataFactory();
        if (null != relator) switch (relator) {
            case "$not":
                OWLObjectComplementOf complement = odf.getOWLObjectComplementOf(axa);
                return complement;
        }
        
        return null;
    }
    
    /**
     * Apply and save an axiom to the configuration ontology
     * @param ax 
     */
    public void applyAxiom(OWLAxiom ax) {
        //If we got an empty axiom back from parsing, we can't add it, so just return
        if (ax == null) {
            return;
        }

        //Reload without configuration
        Ontology.instance().reloadBaseOntologies(true, true);
        OWLOntologyManager mgr = Ontology.instance().getManager();
        OWLOntology ont = Ontology.instance().getOntology(Ontology.instance().getConfig().getPolicyPrefix());
        mgr.addAxiom(ont, ax);

        //Test to see if we're internally consistent, reset if not
        if (!Ontology.instance().isConstistent()) {
            Ontology.instance().reloadBaseOntologies(false, false);  
            this.valid = false;
        } else {
            Ontology.instance().reloadBaseOntologies(true, true);
        }
    }
    
    /**
     * Find out if the current instance is valid
     * @return 
     */
    public boolean isValid() { 
        return this.valid;
    }
    
    /**
     * Get the current explanation if there is one
     * @return 
     */
    public String getExplanation() {
        return this.explanation;
    }
    
    /**
     * Helper function to determine if a triple can construct a valid axiom
     * @param parts
     * @return 
     */
    public boolean isValidAxiom(String[] parts) {
        if (parts.length == 1) {
            //If it is only 1 item, it has to be an equivalent or disjoint class, so handle that
            String subject = parts[0];
            if (!this.valueMap.containsKey(subject)) {
                this.explanation = "value : " + subject + " could not be found in value map for parsing";
                return false;
            }
            subject = this.valueMap.get(subject);
            boolean subjectIsClass = Ontology.instance().classExists(subject);

            if (subjectIsClass) {
                return true;
            } else {
                this.explanation = "value : " + subject + " not a valid Class";
                return false;
            }
        } else if (parts.length == 2) {
            String predicate = parts[0];
            String object = parts[1];
            if (this.expressionMap.containsKey(object) && "$not".equals(predicate)) {
                return true;
            } else {
                if ("$not".equals(predicate)) {
                    this.explanation = "Expression Map doesn't contain : " + object;
                } else {
                    this.explanation = "Invalid predicate : " + predicate;
                }
                return false;
            }
        } else if (parts.length == 3) {
            //Three parts, and we're parsing a standard subject, predicate, object triple
            String subject = parts[0];
            String predicate = parts[1];
            String object = parts[2];

            if (this.valueMap.containsKey(subject) && this.valueMap.containsKey(object)) {
                //For subject, we know we're dealing with classes, object properties, or data properties
                //For object, we know we're dealing with classes, individuals, data literal
                //data literal only occurs if subject is a data property
                        
                subject = this.valueMap.get(subject);
                object = this.valueMap.get(object);

                boolean subjectIsClass = Ontology.instance().classExists(subject);
                boolean subjectIsObjectProperty = Ontology.instance().objectPropertyExists(subject);
                boolean subjectIsDataProperty = Ontology.instance().dataPropertyExists(subject);
                
                boolean objectIsClass = Ontology.instance().classExists(object);
                boolean objectIsIndividual = Ontology.instance().individualExists(object);
                           
                //subject being a class doesn't let us do anything in a 3 part statement since this is 
                //already a class axiom about another class, only works with 1 part
                if (subjectIsClass) {
                    //This isn't a valid statement because it would be covered in 1 part block
                    //Can't relate Object -> Anything other than class
                    this.explanation = "Can't relate Object -> Anything other than class";
                    return false;
                } else if (subjectIsObjectProperty) {
                    if (objectIsClass && "$some".equals(predicate)) {
                        return true;
                    } else if (objectIsIndividual && "$value".equals(predicate)) {
                        return true;                      
                    } else {
                        //If subject is an object property, existing class and individual are all we can choose.
                        //Since we couldn't find it, we have to reject
                        this.explanation = "Subject is object property, and Object isn't class or individual";
                        return false;
                    }
                } else if (subjectIsDataProperty) {
                    //We know that the object has to be a data literal, so we attempt to figure out what it is                  
                    String keyType;
                    String obj = (String) object;
                    
                    if (isInteger(obj)) {
                        keyType = "integer";
                    } else if ("true".equals(obj.toLowerCase()) || "false".equals(obj.toLowerCase())) {
                        keyType = "boolean";
                    } else {
                        keyType = "string";
                    }
                    
                    switch (keyType) {
                        case "string":
                        case "integer":
                        case "boolean":
                            break;
                        default:
                            //Can't process this type for a data literal
                            this.explanation = "Can't convert " + obj + " to valid data literal";
                            return false;
                    }
                    
                    return true; 
                } else {
                    //Not currently in the system.  We have to use the object propery to 
                    //determine what the subject should be, and create it.
                    if (objectIsClass && "$some".equals(predicate.toLowerCase())) {
                        //If the object is a valid Class, then we can assume the subject is an object property
                        return true;
                    } else if (objectIsIndividual && "$value".equals(predicate)) {
                        return true; 
                    } else {
                        //Inadequate information to process
                        this.explanation = "Mismatched predicate with subject/object";
                        return false;
                    }
                }
            } else if (this.expressionMap.containsKey(subject) && this.expressionMap.containsKey(object)) {
                //We're dealing with relational axioms
                return true;
            } else {
                //Not sure, can't evaluate
                this.explanation = "Can't locate variables within value map or expression map, can't parse";
                return false;
            }
        } else {
            //If we have > 3 length, we're dealing with relators and expressions {$and, $or}
            this.expressionMap.put(convertCount(this.count), buildRelationalRestriction(parts));
            return true;
        }    
    }
    
    /**
     * Helper function to build an OWL Axiom from a set of 2 or 3 properties
     * 
     * s = (class | individual | object property | data property)
     * p = ($some (class) | $value (individual) | $and | $or | $not | $nand | $nor)
     * o = (class | individual | object property | data property) 
     * 
     * @param parts
     * @return 
     */
    public boolean buildAxiomFromParts(String[] parts) {
        if (!isValidAxiom(parts)) {
            //Explanation was set if it was invalid, so we can just return false with a more granular response
            return false;
        }
        
        OWLDataFactory odf = Ontology.instance().getDataFactory();  
        
        if (parts.length == 1) {
            String subject = this.valueMap.get(parts[0]);
            IRI subIRI = Ontology.instance().getIRI(subject);
            OWLEntity subEntity = odf.getOWLClass(subIRI);
            OWLClassExpression exp = subEntity.asOWLClass();
            this.expressionMap.put(convertCount(this.count), exp);
            return true;  
        } else if (parts.length == 2) {
            //If we have a 2 part system, we're parsing a not statement
            String predicate = parts[0];
            String object = parts[1];
            this.expressionMap.put(convertCount(this.count), buildRelationalRestriction(predicate, this.expressionMap.get(object)));
            return true;
        } else if (parts.length == 3) {
            //Three parts, and we're parsing a standard subject, predicate, object triple
            String subject = parts[0];
            String predicate = parts[1];
            String object = parts[2];

            if (this.valueMap.containsKey(subject) && this.valueMap.containsKey(object)) {
                //For subject, we know we're dealing with classes, object properties, or data properties
                //For object, we know we're dealing with classes, individuals, data literal
                //data literal only occurs if subject is a data property
                        
                subject = this.valueMap.get(subject);
                object = this.valueMap.get(object);

                boolean subjectIsClass = Ontology.instance().classExists(subject);
                boolean subjectIsObjectProperty = Ontology.instance().objectPropertyExists(subject);
                boolean subjectIsDataProperty = Ontology.instance().dataPropertyExists(subject);
                
                boolean objectIsClass = Ontology.instance().classExists(object);
                boolean objectIsIndividual = Ontology.instance().individualExists(object);
                                            
                IRI subIRI = Ontology.instance().getIRI(subject);
                IRI objIRI;
                OWLEntity subEntity;
                OWLEntity objEntity;
                OWLClassExpression exp;
                
                //subject being a class doesn't let us do anything in a 3 part statement since this is 
                //already a class axiom about another class, only works with 1 part
                if (subjectIsClass) {
                    //This isn't a valid statement because it would be covered in 1 part block
                    //Can't relate Object -> Anything other than class
                    return false;
                } else if (subjectIsObjectProperty) {
                    if (objectIsClass && "$some".equals(predicate)) {
                        objIRI = Ontology.instance().getIRI(object);
                        subEntity = odf.getOWLObjectProperty(subIRI);
                        objEntity = odf.getOWLClass(objIRI);                       
                        exp = odf.getOWLObjectSomeValuesFrom(subEntity.asOWLObjectProperty(), objEntity.asOWLClass());
                        this.expressionMap.put(convertCount(this.count), exp);
                        return true;
                    } else if (objectIsIndividual && "$value".equals(predicate)) {
                        objIRI = Ontology.instance().getIRI(object);
                        subEntity = odf.getOWLObjectProperty(subIRI);
                        objEntity = odf.getOWLNamedIndividual(objIRI);
                        exp = odf.getOWLObjectHasValue(subEntity.asOWLObjectProperty(), objEntity.asOWLNamedIndividual());
                        this.expressionMap.put(convertCount(this.count), exp);
                        return true;                      
                    } else {
                        //If subject is an object property, existing class and individual are all we can choose.
                        //Since we couldn't find it, we have to reject
                        return false;
                    }
                } else if (subjectIsDataProperty) {
                    //We know that the object has to be a data literal, so we attempt to figure out what it is                  
                    OWLLiteral valLiteral = null;
                    String keyType;
                    String obj = (String) object;
                    
                    if (isInteger(obj)) {
                        keyType = "integer";
                    } else if ("true".equals(obj.toLowerCase()) || "false".equals(obj.toLowerCase())) {
                        keyType = "boolean";
                    } else {
                        keyType = "string";
                    }
                    
                    switch (keyType) {
                        case "string":
                            valLiteral = odf.getOWLLiteral((String) object);
                            break;
                        case "integer":
                            valLiteral = odf.getOWLLiteral((int) Integer.parseInt((String) object));
                            break;
                        case "boolean":
                            valLiteral = odf.getOWLLiteral(Boolean.parseBoolean(obj.toLowerCase()));
                            break;
                        default:
                            //Can't process this type for a data literal
                            return false;
                    }
                    
                    subEntity = odf.getOWLDataProperty(subIRI);
                    exp = odf.getOWLDataHasValue(subEntity.asOWLDataProperty(), valLiteral);
                    this.expressionMap.put(convertCount(this.count), exp);
                    return true; 
                } else {
                    //Not currently in the system.  We have to use the object propery to 
                    //determine what the subject should be, and create it.
                    if (objectIsClass) {
                        //If the object is a valid Class, then we can assume the subject is an object property
                        objIRI = Ontology.instance().getIRI(object);
                        subEntity = odf.getOWLObjectProperty(subIRI);
                        objEntity = odf.getOWLClass(objIRI);                       
                        exp = odf.getOWLObjectSomeValuesFrom(subEntity.asOWLObjectProperty(), objEntity.asOWLClass());
                        this.expressionMap.put(convertCount(this.count), exp);
                        return true;
                    } else if (objectIsIndividual) {
                        if ("$value".equals(predicate)) {
                            //If it's a value statement, and object is an individual, we can assume subject 
                            //is an object property.
                            objIRI = Ontology.instance().getIRI(object);
                            subEntity = odf.getOWLObjectProperty(subIRI);
                            objEntity = odf.getOWLNamedIndividual(objIRI);
                            exp = odf.getOWLObjectHasValue(subEntity.asOWLObjectProperty(), objEntity.asOWLNamedIndividual());
                            this.expressionMap.put(convertCount(this.count), exp);
                            return true; 
                        } else {
                            //Don't know how to process this
                            return false; 
                        }
                    } else {
                        //Inadequate information to process
                        return false;
                    }
                }
            } else if (this.expressionMap.containsKey(subject) && this.expressionMap.containsKey(object)) {
                //We're dealing with relational axioms
                this.expressionMap.put(convertCount(this.count), buildRelationalRestriction(predicate, this.expressionMap.get(subject), this.expressionMap.get(object)));
                return true;
            } else {
                //Not sure, can't evaluate
                return false;
            }
        } else {
            //If we have > 3 length, we're dealing with relators and expressions {$and, $or}
            this.expressionMap.put(convertCount(this.count), buildRelationalRestriction(parts));
            return true;
        }       
    }
    
    /**
     * Helper function to determine if a passed value is a valid integer
     * @param s
     * @return 
     */
    public static boolean isInteger(String s) {
        return isInteger(s, 10);
    }

    /**
     * Helper function to determine if a passed value / radix is a valid integer
     * @param s
     * @param radix
     * @return 
     */
    public static boolean isInteger(String s, int radix) {
        if(s.isEmpty()) return false;
        for(int i = 0; i < s.length(); i++) {
            if(i == 0 && s.charAt(i) == '-') {
                if(s.length() == 1) return false;
                else continue;
            }
            if(Character.digit(s.charAt(i),radix) < 0) return false;
        }
        return true;
    }
    
}
