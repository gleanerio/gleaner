# 1. Record architecture decisions

Date: 08-23-2023

## Status

Proposed

## Context

There are some conventions used in the levering of an object store by GleanerIO.
This ADR scopes the naming conventions used both by Gleaner and Nabu.

Some of these conventions have implications on the behavior of the code.  For
example, the URN generation leverages the path structure to establish the 
urn structure (see 0001-URN-decision.md).  

Though the resulting URN is abstracted from the object prefix value, that prefix 
is still used in the initial formation.  

* graphs/
  * graphs/archive
  * graphs/latest
  * graphs/summary
* summoned/
* prov/
* milled/
* orgs/
* reports/
* scheduler/

## Decision
 

## Consequences
 