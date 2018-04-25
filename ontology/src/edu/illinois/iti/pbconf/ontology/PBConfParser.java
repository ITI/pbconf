/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package edu.illinois.iti.pbconf.ontology;

import org.json.JSONObject;

/**
 * Used for parsing content specific to PBConf.  Also handles translation of some properties and operators.
 * @author JD
 */
public class PBConfParser {  
    /**
     * Empty constructor for non-static methods
     */
    public PBConfParser() {}
    
    /**
     * Helper function to process predicate options
     *
     * @param p
     * @return
     */
    public static String translateProperty(String p) {
        String result;

        switch (p) {
            case "ipAddr":
                result = "hasIPAddr";
                break;
            case "routerIPAddr":
                result = "hasRouterIPAddr";
                break;
            case "srcIPAddr":
                result = "hasSrcIPAddr";
                break;
            case "dstIPAddr":
                result = "hasDstIPAddr";
                break;
                
            case "authDNP":
                result = "authDNP";
                break;
            case "authTelnet":
                result = "authTelnet";
                break;
            case "authFTP":
                result = "authFTP";
                break;    

            case "dnp":
            case "dnp3":    
                result = "hasDNP3Stt";
                break;
            case "ftpAnon":
                result = "hasFTPAnonStt";
                break;
            case "ftp":
                result = "hasFTPStt";
                break;
            case "iec":
                result = "hasIEC61850Stt";
                break;
            case "ntp":
                result = "hasNTPStt";
                break;
            case "ping":
                result = "hasPingStt";
                break;
            case "stt":
                result = "hasStt";
                break;
            case "telnet":
                result = "hasTelnetStt";
                break;
            case "accessLogging":
                result = "hasAccessLoggingStt";
                break;
            case "alarm":
                result = "hasAlarmStt";
                break;
            case "gps":
                result = "hasGPSStt";
                break;
            case "role":
                result = "hasRole";
                break;
            case "port5":
                result = "hasPort5";
                break;
            case "port5Cfg":
                result = "hasPort5CfgO";
                break;
            case "sel421Cfg":
                result = "hasSEL421CfgO";
                break;
            case "firewallPolicy":
                result = "hasFirewallPolicy";
                break;
            case "digitalCertificate":
                result = "hasDigitalCertificate";
                break;
            case "certificateAuthority":
                result = "hasCertificateAuthority";
                break;
            case "allowedOperation":
                result = "hasAllowedOperation";
                break;
            case "action":
                result = "hasAction";
                break;

            case "byte":
                result = "hasByte";
                break;
            case "byte1":
                result = "hasByte1";
                break;
            case "byte2":
                result = "hasByte2";
                break;
            case "byte3":
                result = "hasByte3";
                break;
            case "byte4":
                result = "hasByte4";
                break;
            case "order":
                result = "hasOrder";
                break;
            case "port":
                result = "hasPort";
                break;
            case "dstPort":
                result = "hasDstPort";
                break;
            case "srcPort":
                result = "hasStrPort";
                break;
            case "telnetPort":
                result = "hasTelnetPort";
                break;
            case "telnetTimeout":
                result = "hasTelnetTimeout";
                break;
            case "accessTimeout":
                result = "hasAccessTimeout";
                break;
            case "retryDelayed":
                result = "hasRetryDelayed";
                break;
            case "macAddr":
                result = "hasMACAddr";
                break;
            case "expirationDate":
                result = "hasExpirationDate";
                break;    
                
            case "password":
                result = "hasPwd";
                break;
            case "level1":    
            case "password.level1":
                result = "hasLvl1Pwd";
                break;
            case "level2":    
            case "password.level2":
                result = "hasLvl2Pwd";
                break;
            case "levelC":    
            case "password.levelC":
                result = "hasLvlCPwd";
                break;
            case "level1A":    
            case "password.level1A":
                result = "hasLvl1APwd";
                break;
            case "level1B":    
            case "password.level1B":
                result = "hasLvl1BPwd";
                break;
            case "level1O":    
            case "password.level1O":
                result = "hasLvl1OPwd";
                break;
            case "level1P":    
            case "password.level1P":
                result = "hasLvl1PPwd";
                break;

            case "description.type":    
            case "description":    
            case "type":
                result = "type";
                break;

            default:
                result = p;
                break;
        }

        return result;
    }

    /**
     * Determine if we are working with and object property or a data literal
     *
     * @param p
     * @return (object | literal)
     */
    public static String getPropertyType(String p) {
        String result = "";
        switch (p) {
            case "hasSrcIPAddr":
            case "hasDstIPAddr":
            case "hasRouterIPAddr":
            case "hasIPAddr":
                result = "object";
                break;
            case "authDNP":
            case "authTelnet":
            case "authFTP":
                result = "object";
                break;
            case "authusers":
            case "anonymousftp":
                result = "object";
                break;    
            case "hasAllowedOperation":
                result = "object";
                break;
            case "hasDNP3Stt":
            case "hasFTPAnonStt":
            case "hasFTPStt":
            case "hasIEC61850Stt":
            case "hasNTPStt":
            case "hasPingStt":
            case "hasStt":
            case "hasTelnetStt":
            case "hasAccessLoggingStt":
            case "hasAlarmStt":
            case "hasGPSStt":
                result = "object";
                break;
            case "hasRole":
                result = "object";
                break;
            case "hasPort5":
                result = "object";
                break;
            case "hasPort5CfgO":
            case "hasSEL421CfgO":
            case "hasFirewallPolicy":
            case "hasDigitalCertificate":
            case "hasCertificateAuthority":
                result = "object";
                break;
            case "hasAction":
                result = "literal";
                break;
            case "hasByte":
            case "hasByte1":
            case "hasByte2":
            case "hasByte3":
            case "hasByte4":
            case "hasOrder":
            case "hasPort":
            case "hasDstPort":
            case "hasSrcPort":
            case "hasTelnetPort":
            case "hasTelnetTimeout":
            case "hasAccessTimeout":
            case "hasRetryDelayed":
                result = "literal";
                break;
            case "hasMACAddr":
            case "hasPwd":
            case "hasLvl1APwd":
            case "hasLvl1BPwd":
            case "hasLvl1PPwd":
            case "hasLvl1OPwd":
            case "hasLvl1Pwd":
            case "hasLvl2Pwd":
            case "hasLvlCPwd":
                result = "literal";
                break;
            case "hasExpirationDate":
                result = "literal";
                break;
            case "type":
                result = "description";
                break;
            default:
                result = "";
                break;
        }

        return result;
    }

    /**
     * Determine what object or literal range the data has
     *
     * @param p
     * @return (object | literal)
     */
    public static String getPropertyRange(String p) {
        String result = "";
        switch (p) {
            case "hasSrcIPAddr":
            case "hasDstIPAddr":
            case "hasRouterIPAddr":
            case "hasIPAddr":
                result = "IPAddress";
                break;
            case "authDNP":
            case "authTelnet":
            case "authFTP":
                result = "Person";
                break;
            case "authusers":
                result = "Custom";
                break;
            case "anonymousftp":
                result = "Status";
                break;        
            case "hasAllowedOperation":
                result = "SEL421Operation";
                break;
            case "hasDNP3Stt":
            case "hasFTPAnonStt":
            case "hasFTPStt":
            case "hasIEC61850Stt":
            case "hasNTPStt":
            case "hasPingStt":
            case "hasStt":
            case "hasTelnetStt":
            case "hasAccessLoggingStt":
            case "hasAlarmStt":
            case "hasGPSStt":
                result = "Status";
                break;
            case "hasRole":
                result = "Role";
                break;
            case "hasPort5":
                result = "SEL421Port5";
                break;
            case "hasPort5CfgO":
            case "hasSEL421CfgO":
            case "hasFirewallPolicy":
            case "hasDigitalCertificate":
            case "hasCertificateAuthority":
                result = "";
                break;
            case "hasAction":
                result = "boolean";
                break;
            case "hasByte":
            case "hasByte1":
            case "hasByte2":
            case "hasByte3":
            case "hasByte4":
            case "hasOrder":
            case "hasPort":
            case "hasDstPort":
            case "hasSrcPort":
            case "hasTelnetPort":
            case "hasTelnetTimeout":
            case "hasAccessTimeout":
            case "hasRetryDelayed":
                result = "integer";
                break;
            case "hasMACAddr":
            case "hasPwd":
            case "hasLvl1APwd":
            case "hasLvl1BPwd":
            case "hasLvl1PPwd":
            case "hasLvl1OPwd":
            case "hasLvl1Pwd":
            case "hasLvl2Pwd":
            case "hasLvlCPwd":
                result = "string";
                break;
            case "hasExpirationDate":
                result = "string";
                break;
            case "type":
                result = "type";
                break;
            default:
                result = "";
                break;
        }

        return result;
    }
    
    
    /**
     * Based on s, p, o triple, we need to get a breakdown of the axiom we're going to create
     * We also determine if we're using a closed world reasoner, or the built in ontology.
     * 
     * getPropertyRange returns the range (integer, string, boolean, object type, etc)
     * getPropertyType returns the type (object, literal, type)
     * 
     * Here are the basic rules this goes by : 
     * 1.) If subject is empty, it's implied that subject = target.  This is replaced in the parent function.
     * 2.) predicate gives a strong indication of what type of operation we're doing
     * 
     * @param target
     * @param subject
     * @param object
     * @param predicate
     * @return 
     */
    public JSONObject getAxiomParts(String target, String subject, String predicate, Object object) {
        JSONObject json = new JSONObject();

        //getPropertyType
        String objectType = getObjectType(subject, predicate);
        String property = "";
        
        //Allow empty subject's if it's about the target
        if ("".equals(subject)) {
            subject = target;
        }
        
        String objectStr = (String) object;
        String translatedTarget = translateProperty(target);
        String translatedSubject = translateProperty(subject);
        String translatedObject = translateProperty(objectStr);
        
        String policyType = getPolicyType(translatedTarget, translatedSubject, predicate, translatedObject);
        String cwrType = "";
        if (policyType.equals("cwr")) {
            cwrType = getCWRType(translatedTarget, translatedSubject, predicate, translatedObject);
            if (cwrType.equals("") || cwrType.equals("invalid")) {
                policyType = "invalid";
            }
        }
        
        json.put("originalTarget", target);
        json.put("originalSubject", subject);
        json.put("originalPredicate", predicate);
        json.put("originalObject", objectStr);
        
        json.put("target", target);
        json.put("subject", subject);
        json.put("predicate", predicate);
        json.put("object", objectStr);
        
        json.put("translatedTarget", translatedTarget);
        json.put("translatedSubject", translatedSubject);
        json.put("translatedObject", translatedObject);
        
        json.put("policyType", policyType);
        json.put("cwrType", cwrType);
        
        String targetClassification = classifyTarget(translatedTarget, translatedSubject, predicate, translatedObject);
        String subjectClassification = classifySubject(translatedTarget, translatedSubject, predicate, translatedObject);
        String objectClassification = classifyObject(translatedTarget, translatedSubject, predicate, translatedObject);
        
        json.put("targetClassification", targetClassification);
        json.put("subjectClassification", subjectClassification);
        json.put("objectClassification", objectClassification);
        
        json.put("originalValue", (String)object);
        json.put("valueType", objectType);        
        
        switch (objectType) {
            case "integer":
                json.put("value", Integer.parseInt((String)object));
                break;
            case "status":
                if ("on".equals((String)object)) {
                    json.put("value", "<http://iti.illinois.edu/iti/pbconf/core#on>");
                } else {
                    json.put("value", "<http://iti.illinois.edu/iti/pbconf/core#off>");
                }
                break;
            case "string":
            case "IPAddress":     
            default:
                json.put("value", (String)object);
                break;
        }
        
        return json;
    }
    
    /**
     * Classify the target based on the target + S,P,O triple
     * Options are (class | individual | object | data | invalid)
     * invalid indicates that it couldn't be found, or derived
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @return 
     */
    public String classifyTarget(String target, String subject, String predicate, String object) { 
        boolean targetIsClass = Ontology.instance().classExists(target);
        boolean targetIsIndividual = Ontology.instance().individualExists(target);
        boolean targetIsDataProperty = Ontology.instance().dataPropertyExists(target);
        boolean targetIsObjectProperty = Ontology.instance().objectPropertyExists(target);
        boolean targetIsUnknown = (!targetIsClass && !targetIsIndividual && !targetIsDataProperty && !targetIsObjectProperty);
        
        boolean subjectIsClass = Ontology.instance().classExists(subject);
        boolean subjectIsIndividual = Ontology.instance().individualExists(subject);
        boolean subjectIsDataProperty = Ontology.instance().dataPropertyExists(subject);
        boolean subjectIsObjectProperty = Ontology.instance().objectPropertyExists(subject);
        boolean subjectIsUnknown = (!subjectIsClass && !subjectIsIndividual && !subjectIsDataProperty && !subjectIsObjectProperty);

        boolean objectIsClass = Ontology.instance().classExists(object);
        boolean objectIsIndividual = Ontology.instance().individualExists(object);
        boolean objectIsDataProperty = Ontology.instance().dataPropertyExists(object);
        boolean objectIsObjectProperty = Ontology.instance().objectPropertyExists(object);
        
        if (targetIsUnknown) {
            /**
             * This is the most important / difficult case as we're trying to impose policy on something
             * when we don't know what we're actually referring to.  We can try to ascertain that from what
             * policy is being set.  If it's completely ambiguous, we have to reject it.
             * 
             * Right now, we're not going to factor in new classes since we should be sticking with what we 
             * currently support (SEL421, LINUX), unless target==subject, which should only happen in that case
             */
            if (subjectIsUnknown) {
                //Check to see if target == subject, but not empty
                if (target.equals(subject) && !"".equals(target)) {
                    //We're talking about an unreferenced class or individual
                    if (objectIsDataProperty || objectIsObjectProperty) {
                        return "individual";
                    } else if (objectIsClass || objectIsIndividual) {
                        return "class";
                    } else {
                        return "invalid";
                    }
                } else {
                    return "invalid";
                }
            } else if (subjectIsClass) {
                //This is a weird situation, but just in case
                return "class";
            } else if (subjectIsIndividual) {
                //This case shouldn't occur, so we'll return invalid
                return "invalid";
            } else if (subjectIsObjectProperty || subjectIsDataProperty) {
                //If the subject is an object or data property, we can assume we're talking about an individual
                return "individual";
            } else {
                return "invalid";
            }
        } else {
            if (targetIsClass) {
                return "class";
            } else if (targetIsIndividual) {
                return "individual";
            } else if (targetIsObjectProperty) {
                return "object";
            } else if (targetIsDataProperty) {
                return "data";
            } else {
                return "invalid";
            }
        }
    }
    
    /**
     * Classify the subject based on the target + S,P,O triple
     * Options are (class | individual | object | data) 
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @return 
     */
    public String classifySubject(String target, String subject, String predicate, String object) {
        boolean targetIsClass = Ontology.instance().classExists(target);
        boolean targetIsIndividual = Ontology.instance().individualExists(target);
        boolean targetIsDataProperty = Ontology.instance().dataPropertyExists(target);
        boolean targetIsObjectProperty = Ontology.instance().objectPropertyExists(target);
   
        boolean subjectIsClass = Ontology.instance().classExists(subject);
        boolean subjectIsIndividual = Ontology.instance().individualExists(subject);
        boolean subjectIsDataProperty = Ontology.instance().dataPropertyExists(subject);
        boolean subjectIsObjectProperty = Ontology.instance().objectPropertyExists(subject);
        boolean subjectIsUnknown = (!subjectIsClass && !subjectIsIndividual && !subjectIsDataProperty && !subjectIsObjectProperty);

        boolean objectIsClass = Ontology.instance().classExists(object);
        boolean objectIsIndividual = Ontology.instance().individualExists(object);
        boolean objectIsDataProperty = Ontology.instance().dataPropertyExists(object);
        boolean objectIsObjectProperty = Ontology.instance().objectPropertyExists(object);
        boolean objectIsUnknown = (!objectIsClass && !objectIsIndividual && !objectIsDataProperty && !objectIsObjectProperty);
         
        if (subjectIsUnknown) {
            if ("".equals(subject) || (subject.equals(target) && !"".equals(subject))) {
                //If the subject is empty, we can just return the target type
                if (targetIsClass) {
                    return "class";
                } else if (targetIsIndividual) {
                    return "individual";
                } else if (targetIsObjectProperty) {
                    return "object";
                } else if (targetIsDataProperty) {
                    return "data";
                } else {
                    switch (predicate) {
                        case "requires":
                        case "capability":
                        case "mustNotHave":
                            if (objectIsDataProperty || objectIsObjectProperty) {
                                return "individual";
                            } else if (objectIsClass || objectIsIndividual) {
                                return "class";
                            } else {
                                return "invalid";
                            }
                        case "isA":
                        case "isNotA":
                            return "individual";
                        case "state":
                        case "status":
                            //State or status is setting {on, off}
                            //This only occurs for object properties, so return "object"
                            return "object";
                        case "eq":
                        case "neq":    
                        case "gt":
                        case "lt":
                        case "gte":
                        case "lte":
                            return "data";
                        default:
                            if (objectIsUnknown && !"".equals(object)) {
                                //Should imply a data literal
                                return "data";
                            } else {
                                return "invalid";
                            }
                    }    
                }
            } else {
                //At this point, subject is non-empty, is unknown, 
                //but doesn't match target, so we try to derive it from predicate + object
                switch (predicate) {
                    case "requires":
                    case "capability":
                    case "mustNotHave":
                        if (objectIsDataProperty || objectIsObjectProperty) {
                            return "individual";
                        } else if (objectIsClass || objectIsIndividual) {
                            return "class";
                        } else {
                            return "invalid";
                        }
                    case "isA":
                    case "isNotA":
                        return "individual";
                    case "state":
                    case "status":
                        //State or status is setting {on, off}
                        //This only occurs for object properties, so return "object"
                        return "object";
                    case "eq":
                    case "neq":    
                    case "gt":
                    case "lt":
                    case "gte":
                    case "lte":
                        return "data";
                    default:
                        if (objectIsUnknown && !"".equals(object)) {
                            //Should imply a data literal
                            return "data";
                        } else {
                            return "invalid";
                        }
                }
            }
        } else if (subjectIsClass) {
            //This is a weird situation, but just in case
            return "class";
        } else if (subjectIsIndividual) {
            //This case shouldn't occur, so we'll return invalid
            return "individual";
        } else if (subjectIsObjectProperty) {
            //If the subject is an object or data property, we can assume we're talking about an individual
            return "object";
        } else if (subjectIsDataProperty) {
            return "data";
        } else {
            return "invalid";
        }
    }
    
    /**
     * Classify the object based on the target + S,P,O triple
     * Options are (class | individual | object | data) 
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @return 
     */
    public String classifyObject(String target, String subject, String predicate, String object) {
        String subjectType = classifySubject(target, subject, predicate, object);
        String targetType = classifyTarget(target, subject, predicate, object);
              
        boolean objectIsClass = Ontology.instance().classExists(object);
        boolean objectIsIndividual = Ontology.instance().individualExists(object);
        boolean objectIsDataProperty = Ontology.instance().dataPropertyExists(object);
        boolean objectIsObjectProperty = Ontology.instance().objectPropertyExists(object);
        boolean objectIsUnknown = (!objectIsClass && !objectIsIndividual && !objectIsDataProperty && !objectIsObjectProperty);
        
        if (objectIsUnknown) {
            if ("invalid".equals(subjectType) || "invalid".equals(targetType)) {
                //If either one of these is invalid, and object is unknown, we don't know enough to say for sure
                return "invalid";
            }
            
            switch (subjectType) {
                case "data":
                    return "literal";
                case "object":
                    return "individual";
                case "individual":
                    //subject shouldn't be individual
                    return "invalid";
                case "class":
                    //If it's a class, then it's dependent on the predicate
                    switch (predicate) {
                        case "requires":
                        case "capability":
                        case "mustNotHave":
                            //Can't be sure if it's object or data property
                            return "invalid";
                        case "isA":
                        case "isNotA":
                            //When predicate is one of these, it is a class
                            return "class";
                        case "state":
                        case "status":
                            //State or status is setting {on, off}
                            //This only occurs for object properties, so return "object"
                            return "individual";
                        case "eq":
                        case "neq":
                        case "gt":
                        case "lt":
                        case "gte":
                        case "lte":
                            return "literal";
                        default:
                            return "invalid";
                    }
                default:
                    return "invalid";
            }
        } else {
            if (objectIsClass) {
                return "class";
            } else if (objectIsIndividual) {
                return "individual"; 
            } else if (objectIsObjectProperty) {
                return "object";
            } else if (objectIsDataProperty) {
                return "data";
            } else if (objectIsUnknown) {
                return "literal";
            } else {
                return "invalid";
            }
        }
    }
    
    /**
     * Predicate will determine what type of command we're trying to process,
     * Target will determine what type of reasoner we're trying to use.
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @return 
     */
    public String getPolicyType(String target, String subject, String predicate, String object) { 
        String targetType = classifyTarget(target, subject, predicate, object);
        String subjectType = classifySubject(target, subject, predicate, object);
        String objectType = classifyObject(target, subject, predicate, object);
        
        //If any of these are invalid, the policy itself is invalid
        if ("invalid".equals(targetType) || "invalid".equals(subjectType) || "invalid".equals(objectType)) {
            return "invalid";
        }
        
        //We now know what we can ascertain about target, subject, and object, which should give us enough
        //to say what type of reasoning we need to use
        String policyType = "";
        switch (targetType) {
            //Here, we're discussing policy that is being set regarding a class
            case "class":
                //Most of the class related stuff would have been handled at a higher level, so some 
                //of it will just be considered invalid at this point
                switch (objectType) {
                    case "class":
                        //This should only really come about if we're using 'isA', 'isNotA'
                        policyType = "cwr";
                        break;
                    case "individual":
                        //This shouldn't happen because we don't currently offer a way to compare
                        //these values that wouldn't have been caught earlier
                        policyType = "cwr";
                        break;
                    case "object":
                        //this like likely a requires request, or a property requirements
                        policyType = "cwr";
                        break;
                    case "data":
                        //this like likely a requires request, or a property requirements
                        policyType = "cwr";
                        break;
                    case "literal":
                        //this is likely a property requirements
                        policyType = "cwr";
                        break;
                    default:
                        policyType = "invalid";
                        break;
                }
                break;
            //Next, we're on to policy related to a specific individual    
            case "individual":
                switch (objectType) {
                    case "class":
                        //This shouldn't happen because we don't currently offer a way to compare
                        //these values that wouldn't have been caught earlier (equivalentTo, disjointWith)
                        policyType = "cwr";
                        break;
                    case "individual":
                        //This shouldn't happen because we don't currently offer a way to compare
                        //these values that wouldn't have been caught earlier
                        policyType = "cwr";
                        break;
                    case "object":
                        //this like likely a requires request, or a property requirements
                        policyType = "cwr";
                        break;
                    case "data":
                        //this like likely a requires request, or a property requirements
                        policyType = "cwr";
                        break;
                    case "literal":
                        //this is likely a property requirements
                        policyType = "cwr";
                        break;
                    default:
                        policyType = "invalid";
                        break;
                }
                break;
            //Handle situation involving object property policy    
            case "object":
                switch (objectType) {
                    case "class":
                        policyType = "invalid";
                        break;
                    case "individual":
                        policyType = "invalid";
                        break;
                    case "object":
                        policyType = "invalid";
                        break;
                    case "data":
                        policyType = "invalid";
                        break;
                    case "literal":
                        policyType = "invalid";
                        break;
                    default:
                        policyType = "invalid";
                        break;
                }
                break;
            //Handle policy for data properties    
            case "data":
                switch (objectType) {
                    case "class":
                        //This shouldn't happen because we don't currently offer a way to compare
                        //these values that wouldn't have been caught earlier (equivalentTo, disjointWith)
                        policyType = "invalid";
                        break;
                    case "individual":
                        //This shouldn't happen because we don't currently offer a way to compare
                        //these values that wouldn't have been caught earlier
                        policyType = "invalid";
                        break;
                    case "object":
                        //this like likely a requires request, or a property requirements
                        policyType = "invalid";
                        break;
                    case "data":
                        //this like likely a requires request, or a property requirements
                        policyType = "cwr";
                        break;
                    case "literal":
                        //this is likely a property requirements
                        policyType = "data";
                        break;
                    default:
                        policyType = "invalid";
                        break;
                }
                break;
            //This is of unknown type, just mark as invalid    
            default:
                policyType = "invalid";
                break;
                
        }
        
        return policyType;
    }
    
    /**
     * Predicate will determine what type of command we're trying to process,
     * Target will determine what type of reasoner we're trying to use.
     * @param target
     * @param subject
     * @param predicate
     * @param object
     * @return 
     */
    public String getCWRType(String target, String subject, String predicate, String object) { 
        String targetType = classifyTarget(target, subject, predicate, object);
        String subjectType = classifySubject(target, subject, predicate, object);
        String objectType = classifyObject(target, subject, predicate, object);
        
        String result = "";
        
        //If any of these are invalid, the policy itself is invalid
        if ("invalid".equals(targetType) || "invalid".equals(subjectType) || "invalid".equals(objectType)) {
            result = "invalid";
        }
        
        switch (targetType.toLowerCase()) {
            case "class":
                switch (predicate.toLowerCase()) {
                    case "requires":
                    case "capability":
                        result = "ClassMustHaveProperty";
                        break;
                    case "mustnothave":
                        result = "ClassMustNotHaveProperty";
                        break;
                    case "state":
                    case "status":
                    case "isa":
                    case "isnota":
                    case "gt":
                    case "lt":
                    case "gte":
                    case "lte":
                    case "eq":
                    case "neq":
                    case "min-length":
                    case "max-length":
                    case "complexity":     
                        result = "ClassPropertyRestrictions";
                        break;
                    default:
                        break;
                }   break;
            case "individual":
                switch (predicate.toLowerCase()) {
                    case "requires":
                    case "capability":
                        result = "IndividualMustHaveProperty";
                        break;
                    case "mustnothave":
                        result = "IndividualMustNotHaveProperty";
                        break;
                    case "state":
                    case "status":
                    case "isa":
                    case "isnota":
                    case "gt":
                    case "lt":
                    case "gte":
                    case "lte":
                    case "eq":
                    case "neq":
                    case "min-length":
                    case "max-length":
                    case "complexity":  
                        result = "IndividualPropertyRestrictions";
                        break;
                default:
                    break;    
            }   break;
            default:
                result = "invalid";
                break;
        }
        
        return result;
    }
    
    /**
     * Determine the type of an object based on the predicate passed
     * @param subject
     * @param predicate
     * @return 
     */
    public String getObjectType(String subject, String predicate) {
        String result = "";
        
        if ("".equals(subject)) {
            if ("min-length".equals(predicate) || "max-length".equals(predicate)) {
                result = "integer";
            }
            if ("complexity".equals(predicate)) {
                result = "string";
            }
            if ("disjointWith".equals(predicate)) {
                result = "string";
            }
        } else if (subject.startsWith("password")) {
            if ("min-length".equals(predicate) || "max-length".equals(predicate)) {
                result = "integer";
            }
            if ("complexity".equals(predicate)) {
                result = "string";
            }
        } else {
            switch (subject) {  
                case "macAddr":
                    result = "string";
                    break;    
                case "accessTimeout":
                    result = "integer";
                    break;
                case "telnetPort":
                    result = "integer";
                    break; 
                case "telnetTimeout":
                    result = "integer";
                    break;
                case "ntp":
                    result = "status";
                    break;
                case "ftp":
                    result = "status";
                    break;
                case "telnet":
                    result = "status";
                    break;
                case "ping":
                    result = "status";
                    break;
                case "dnp":
                case "dnp3":    
                    result = "status";
                    break;
                case "iec":
                    result = "status";
                    break;
                case "ipAddr":
                    result = "IPAddress";
                    break;
                case "routerIPAddr":
                    result = "IPAddress";
                    break;      
                default:
                    result = "";
                    break;
            }
        }
        
        //Fall back to a string value if we can't find it
        if ("".equals(result)) {
            return "string";
        } else {
            return result;
        }
    }
    
}
