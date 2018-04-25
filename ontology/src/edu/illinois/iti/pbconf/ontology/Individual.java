/*
 *
 */
package edu.illinois.iti.pbconf.ontology;

import java.util.HashSet;
import java.util.Map;
import java.util.Set;
import org.apache.log4j.Logger;
import org.semanticweb.owlapi.model.AddAxiom;
import org.semanticweb.owlapi.model.IRI;
import org.semanticweb.owlapi.model.OWLClass;
import org.semanticweb.owlapi.model.OWLClassAssertionAxiom;
import org.semanticweb.owlapi.model.OWLDataFactory;
import org.semanticweb.owlapi.model.OWLDataProperty;
import org.semanticweb.owlapi.model.OWLDataPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLDataPropertyExpression;
import org.semanticweb.owlapi.model.OWLIndividual;
import org.semanticweb.owlapi.model.OWLIndividualAxiom;
import org.semanticweb.owlapi.model.OWLLiteral;
import org.semanticweb.owlapi.model.OWLNamedIndividual;
import org.semanticweb.owlapi.model.OWLObjectProperty;
import org.semanticweb.owlapi.model.OWLObjectPropertyAssertionAxiom;
import org.semanticweb.owlapi.model.OWLObjectPropertyExpression;
import org.semanticweb.owlapi.model.OWLOntology;
import org.semanticweb.owlapi.model.OWLOntologyManager;
import org.semanticweb.owlapi.model.RemoveAxiom;
import org.semanticweb.owlapi.reasoner.NodeSet;

/**
 * A class to simplify typical operations on ontology individuals. An Individual
 * instance is typically obtained from the factory method
 * Ontology.getIndividual().
 *
 * @author anderson
 */
public class Individual {

    static Logger logger = Logger.getLogger(Individual.class.getName().replaceFirst(".+\\.", ""));

    IRI nameIRI;
    OWLNamedIndividual self;
    OWLOntology ontology;

//    protected Individual(IRI nameIRI) {
//        this(nameIRI,Ontology.instance().getRootOntology());
//    }
    /**
     * New individuals should be created by Ontology.getIndividual()
     *
     * @param nameIRI
     * @param ontology
     */
    protected Individual(IRI nameIRI, OWLOntology ontology) {
        this.nameIRI = nameIRI;
        this.ontology = ontology;

        OWLNamedIndividual ind = Ontology.instance().getDataFactory().getOWLNamedIndividual(nameIRI);
        self = ind;
        // previously, an OWLNamedIndividual straight from the data factory would'nt
        // find data properties that are obviously right there in the ontology,
        // but an OWLNamedIndividual from the reasoner works. Why?

        // !!! turns out that the getIRI(prefix,suffix) returns a broken IRI. or so it seems.
        // Therefore, if we can get a same as individual from the reasoner, prefer that.
        // see IndividualTest.test_debugDataProperties
//        Set<OWLNamedIndividual> sames = Ontology.instance().getReasoner().getSameIndividuals(ind).getEntities();
//        OWLNamedIndividual same = null;
//        for (OWLNamedIndividual _same : sames) {
//            if (_same.getIRI().toString().equals(ind.getIRI().toString())) {
//                same = _same;
//                break;
//            }
//        }
//        if (same != null) {
//            self = same;
//        }
//        else {
//            // use the one from factory
//            self = ind;
//        }
    }

    /**
     * Get all classes for Individual
     *
     * @param individualName
     * @return
     */
    public Set<OWLClass> getClasses(String individualName) {
        Ontology ont = Ontology.instance();
        NodeSet<OWLClass> classNodeSet = ont.getReasoner().getTypes(self, true);
        Set<OWLClass> classes = classNodeSet.getFlattened();
        return classes;
    }
    
    /**
     * For those times when the simplified Individual interface is not enough.
     *
     * @return
     */
    public OWLNamedIndividual getOWLIndividual() {
        return self;
    }

    /**
     * Individual IRI
     *
     * @return IRI
     */
    public IRI getIRI() {
        return nameIRI;
    }

    /**
     * Return simplified name for Individual.
     *
     * @return String
     */
    public String getName() {
        return Ontology.instance().getSimpleName(nameIRI);
    }

    /**
     * Set a generalized object data property (determine based on value) Saves
     * us some difficulties for parsing json objects Hides some of the
     * hideousness of OWL API.
     *
     * @param property
     * @param value
     */
    public void setProperty(IRI property, Object value) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral valLiteral = null;

        if (value.getClass().equals(Integer.class)) {
            valLiteral = dataFactory.getOWLLiteral((int) value);
        } else if (value.getClass().equals(String.class)) {
            valLiteral = dataFactory.getOWLLiteral((String) value);
        } else if (value.getClass().equals(Boolean.class)) {
            valLiteral = dataFactory.getOWLLiteral((Boolean) value);
        }

        setProperty(property, valLiteral);
    }

    /**
     * Set a String data property. Hides some of the hideousness of OWL API.
     *
     * @param property
     * @param value
     */
    public void setProperty(IRI property, String value) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral stringLiteral = dataFactory.getOWLLiteral(value);
        setProperty(property, stringLiteral);
    }

    /**
     * Set an int property.
     *
     * @param property
     * @param value
     */
    public void setProperty(IRI property, int value) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral stringLiteral = dataFactory.getOWLLiteral(value);
        setProperty(property, stringLiteral);
    }

    /**
     * Set a double property.
     *
     * @param property
     * @param value
     */
    public void setProperty(IRI property, double value) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral stringLiteral = dataFactory.getOWLLiteral(value);
        setProperty(property, stringLiteral);
    }

    /**
     * Set a boolean property.
     *
     * @param property
     * @param value
     */
    public void setProperty(IRI property, boolean value) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLLiteral stringLiteral = dataFactory.getOWLLiteral(value);
        setProperty(property, stringLiteral);
    }

    /**
     * Set a data property. The property axiom will be added to the Individuals
     * home ontology (from ctor)
     *
     * @param property
     * @param object
     */
    public void setProperty(IRI property, OWLLiteral object) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLOntologyManager manager = Ontology.instance().getManager();
        //OWLOntology baseOnt = Ontology.instance().getRootOntology();
        OWLIndividual self = getOWLIndividual();
        OWLDataProperty prop = dataFactory.getOWLDataProperty(property);
        
        OWLDataPropertyAssertionAxiom axiom = dataFactory.getOWLDataPropertyAssertionAxiom(prop, self, object);
        logger.debug(axiom);
        manager.applyChange(new AddAxiom(ontology, axiom));
    }
    
    /**
     * Clear an existing configuration setting
     * @param property 
     */
    public void clearProperty(IRI property) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLOntologyManager manager = Ontology.instance().getManager();
        OWLIndividual ind = getOWLIndividual();
        OWLDataProperty prop = dataFactory.getOWLDataProperty(property);
        Set<OWLLiteral> val = getDataProperty(property);
        
        for (OWLLiteral v : val) {
            OWLDataPropertyAssertionAxiom oldAxiom = dataFactory.getOWLDataPropertyAssertionAxiom(prop, ind, v);           
            manager.applyChange(new RemoveAxiom(ontology, oldAxiom));
        }     
    }

    /**
     * Set an object property. The property axiom will be added to the
     * Individuals home ontology (from ctor)
     *
     * @param property
     * @param object
     */
    public void setProperty(IRI property, Individual object) {
        setProperty(property, object.getOWLIndividual());
    }

    /**
     * Set an object property. The property axiom will be added to the
     * Individuals home ontology (from ctor)
     *
     * @param property
     * @param object
     */
    public void setProperty(IRI property, OWLIndividual object) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLOntologyManager manager = Ontology.instance().getManager();
        //OWLOntology baseOnt = Ontology.instance().getRootOntology();
        OWLIndividual self = getOWLIndividual();
        OWLObjectPropertyExpression objProp = dataFactory.getOWLObjectProperty(property);
        OWLObjectPropertyAssertionAxiom axiom
                = dataFactory.getOWLObjectPropertyAssertionAxiom(objProp, self, object);
        manager.applyChange(new AddAxiom(ontology, axiom));
    }

    /**
     * Get all values of a data property. Returns all reasoned values of data
     * property. This implies that values will be drawn from the root ontology,
     * and all of its imported ontologies, and presumes that the home ontology
     * of this Individual is imported into root.
     *
     * @param property
     * @return
     */
    public Set<OWLLiteral> getDataProperty(IRI property) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLDataProperty prop = dataFactory.getOWLDataProperty(property);
        // this should be across all linked ontologies.
        return Ontology.instance().getReasoner().getDataPropertyValues(self, prop);
    }

    /**
     * Get all values of an object property. Returns all reasoned values of
     * object property.
     *
     * @param property
     * @return
     */
    public Set<OWLNamedIndividual> getObjectProperty(IRI property) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLObjectPropertyExpression objProp = dataFactory.getOWLObjectProperty(property);
        NodeSet<OWLNamedIndividual> nodes = Ontology.instance().getReasoner().getObjectPropertyValues(self, objProp);
        return nodes.getFlattened();
    }

    /**
     * Get the data properties of this individual. You can get the values of the
     * data properties with getDataProperty.
     *
     * @return
     */
    public Set<OWLDataProperty> getDataProperties() {
        OWLOntology baseOnt = Ontology.instance().getRootOntology();
        OWLIndividual self = getOWLIndividual();

        Set<OWLDataProperty> properties = new HashSet<>();
        Set<OWLOntology> onts = baseOnt.getImportsClosure();
        for (OWLOntology ont : onts) {
            Map<OWLDataPropertyExpression, Set<OWLLiteral>> dataProps = self.getDataPropertyValues(ont);
            for (OWLDataPropertyExpression exp : dataProps.keySet()) {
                if (!exp.isAnonymous()) {
                    OWLDataProperty dp = exp.asOWLDataProperty();
                    properties.add(dp);
                }
            }
        }

        return properties;
    }

    /**
     * Get all object properties.
     * @return object properties
     */
    public Set<OWLObjectProperty> getObjectProperties() {
        OWLOntology baseOnt = Ontology.instance().getRootOntology();
        OWLIndividual self = getOWLIndividual();

        Set<OWLObjectProperty> properties = new HashSet<>();
        Set<OWLOntology> onts = baseOnt.getImportsClosure();
        for (OWLOntology ont : onts) {
            Map<OWLObjectPropertyExpression, Set<OWLIndividual>> objProps = self.getObjectPropertyValues(ont);
            for (OWLObjectPropertyExpression exp : objProps.keySet()) {
                if (!exp.isAnonymous()) {
                    OWLObjectProperty op = exp.asOWLObjectProperty();
                    properties.add(op);
                }
            }
        }

        return properties;
    }

    /**
     * Get all axioms of the individual, 
     * @return 
     */
    public Set<OWLIndividualAxiom> getAxioms() {
        OWLOntology baseOnt = Ontology.instance().getRootOntology();
        OWLIndividual self = getOWLIndividual();

        Set<OWLIndividualAxiom> axioms = new HashSet<>();
        Set<OWLOntology> onts = baseOnt.getImportsClosure();
        for (OWLOntology ont : onts) {
            Set<OWLIndividualAxiom> _axioms = ont.getAxioms(self);
            axioms.addAll(_axioms);
        }

        return axioms;
    }
    /**
     * set class. Note that this adds a new class assertion. Could be
     * "addClass".
     *
     * @param classIRI
     */
    public void setClass(IRI classIRI) {
        OWLDataFactory dataFactory = Ontology.instance().getDataFactory();
        OWLIndividual ind = getOWLIndividual();
        OWLClass cls = dataFactory.getOWLClass(classIRI);
        OWLClassAssertionAxiom classAssertion = dataFactory.getOWLClassAssertionAxiom(cls, ind);
        setClass(classAssertion);
    }

    /**
     * set class.
     *
     * @param isA
     */
    public void setClass(OWLClassAssertionAxiom isA) {
        OWLOntologyManager manager = Ontology.instance().getManager();
        manager.applyChange(new AddAxiom(ontology, isA));
    }

}
