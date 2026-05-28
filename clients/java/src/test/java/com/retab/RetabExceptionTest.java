// Regression test for the typed exception hierarchy emitted by oagen.
// Confirms that RetabException extends IOException (so `throws IOException`
// keeps working) and that fromStatusCode returns the matching subclass.

package com.retab;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import org.junit.jupiter.api.Test;

class RetabExceptionTest {
  @Test
  void baseExceptionIsIoException() {
    RetabException e = new RetabException("boom", 500, "{}");
    assertInstanceOf(IOException.class, e);
    assertEquals(500, e.getStatusCode());
    assertEquals("{}", e.getResponseBody());
  }

  @Test
  void fromStatusCodeReturnsTypedSubclass() {
    assertInstanceOf(
        RetabAuthenticationException.class, RetabException.fromStatusCode(401, "bad key", null));
    assertInstanceOf(
        RetabPermissionException.class, RetabException.fromStatusCode(403, "no perm", null));
    assertInstanceOf(
        RetabNotFoundException.class, RetabException.fromStatusCode(404, "missing", null));
    assertInstanceOf(
        RetabRateLimitException.class, RetabException.fromStatusCode(429, "slow down", null));
  }

  @Test
  void fromStatusCodeFallsBackToBaseForUnknownStatus() {
    RetabException e = RetabException.fromStatusCode(500, "server", null);
    assertEquals(RetabException.class, e.getClass());
    assertEquals(500, e.getStatusCode());
  }

  @Test
  void subclassesAreCatchableAsIoException() {
    boolean caught = false;
    try {
      throw RetabException.fromStatusCode(404, "missing", null);
    } catch (IOException ignored) {
      caught = true;
    }
    assertTrue(caught, "RetabNotFoundException should be catchable via IOException");
  }
}
